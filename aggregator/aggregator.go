package aggregator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/docker/docker/client"
	"github.com/gorilla/feeds"
	log "github.com/sirupsen/logrus"
	"github.com/otiai10/copy"
)

type Aggregator struct {
	Config  *Config
	Content map[string]*feeds.Feed
	docker  *client.Client
}

func NewAggregator(configPath string) (a *Aggregator, err error) {

	a = &Aggregator{
		Config:  &Config{},
		Content: map[string]*feeds.Feed{},
	}
	if err := a.LoadConfig(configPath); err != nil {
		return a, err
	}

	a.docker, err = client.NewClientWithOpts()
	if err != nil {
		return a, err
	}

	return a, client.FromEnv(a.docker)
}

func (a *Aggregator) LoadConfig(configPath string) error {

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, a.Config); err != nil {
		return err
	}

	if a.Config.TimeZone == "" {
		a.Config.TimeZone = "UTC"
	}

	if a.Config.StaticAssets == "" {
		a.Config.StaticAssets = "static"
	}

	return nil
}

func (a *Aggregator) Start() error {

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, os.Kill, syscall.SIGHUP, syscall.SIGKILL, syscall.SIGQUIT)


	ticker := time.NewTicker(time.Second * time.Duration(a.Config.RefreshIntervalSeconds))
	if err := a.Run(); err != nil {
		return err
	}
	for {
		select {
		case <-c:
			// TODO shutdown docker container
			return nil
		case <-ticker.C:
			if err := a.Run(); err != nil {
				return err
			}
		}
	}
}

func (a *Aggregator) Run() error {

	if err := a.StartDocker(); err != nil {
		return err
	}

	if err := a.Fetch(); err != nil {
		return err
	}

	for _, dir := range []string{"rss", "atom", "json"} {
		path := fmt.Sprintf("%s/%s", a.Config.OutputPath, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("could not create path %s, %s", path, err.Error())
		}
	}

	if err := copy.Copy(a.Config.StaticAssets, a.Config.OutputPath); err != nil {
		return fmt.Errorf("could not copy static assets, %s", err.Error())
	}

	if err := a.WriteHTML(); err != nil {
		return err
	}

	if err := a.WriteRSS(); err != nil {
		return err
	}

	if a.Config.AutoCommit {
		if err := a.Commit(); err != nil {
			return err
		}
	}

	log.Info("static assets updated")

	return nil
}