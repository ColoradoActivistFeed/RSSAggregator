package aggregator

import (
	"bufio"
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	log "github.com/sirupsen/logrus"
)

const containerName = "rss-aggregator-rss-bridge"
const imageName = "rssbridge/rss-bridge:latest" // TODO pin version

type dockerPullProgress struct {
	Status         string `json:"status"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
	Id string `json:"id"`
}

func (a *Aggregator) StartDocker() error {

	id, running, err := a.FindDockerContainer()
	if err != nil {
		return err
	}

	if id != "" {
		if running {
			if err := a.StopDockerContainer(id); err != nil {
				return err
			}
		}
		if err := a.RemoveDockerContainer(id); err != nil {
			return err
		}
		log.WithField("id", id).Info("removed old docker container")
	}

	if err := a.PullDockerImage(); err != nil {
		return err
	}
	log.WithField("image", imageName).Info("pulled image")

	id, err = a.CreateDockerContainer()
	if err != nil {
		return err
	}
	log.WithField("id", id).Info("created docker container")

	if err := a.StartDockerContainer(id); err != nil {
		return err
	}
	log.WithField("id", id).Info("started docker container")

	// TODO check for process ready
	time.Sleep(time.Second * 5)

	return nil
}

func (a *Aggregator) FindDockerContainer() (string, bool, error) {

	containers, err := a.docker.ContainerList(
		context.Background(),
		types.ContainerListOptions{
			All: true,
		},
	)
	if err != nil {
		return "", false, err
	}

	for _, c := range containers {
		for _, name := range c.Names {
			if strings.HasPrefix(name, "/") {
				name = name[1:]
			}
			if name == containerName {
				return c.ID, c.State == "running", nil
			}
		}
	}

	return "", false, nil
}

func (a *Aggregator) StartDockerContainer(id string) (err error) {
	return a.docker.ContainerStart(context.Background(), id, types.ContainerStartOptions{})
}

func (a *Aggregator) RemoveDockerContainer(id string) (err error) {
	return a.docker.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{})
}

func (a *Aggregator) StopDockerContainer(id string) (err error) {
	timeout := time.Second * 5
	return a.docker.ContainerStop(context.Background(), id, &timeout)
}

func (a *Aggregator) PullDockerImage() (err error) {

	r, err := a.docker.ImagePull(context.Background(), imageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		var status dockerPullProgress
		if err := json.Unmarshal(scanner.Bytes(), &status); err != nil {
			return err
		}

		l := log.NewEntry(log.StandardLogger())
		if status.Id != "" {
			l = l.WithField("id", status.Id)
		}
		if status.ProgressDetail.Total != 0 {
			l = l.WithField("total", status.ProgressDetail.Total)
		}
		if status.ProgressDetail.Current != 0 {
			l = l.WithField("current", status.ProgressDetail.Current)
		}
		l.Info(status.Status)
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return
}

func (a *Aggregator) CreateDockerContainer() (id string, err error) {

	r, err := a.docker.ContainerCreate(context.Background(),
		&dockerContainer.Config{
			Image: imageName,
		},
		&dockerContainer.HostConfig{
			PortBindings: map[nat.Port][]nat.PortBinding{
				"80/tcp": {
					{HostIP: "127.0.0.1", HostPort: "1200"},
				},
			},
			RestartPolicy: dockerContainer.RestartPolicy{
				Name: "always",
			},
		},
		nil,
		containerName,
	)
	if err != nil {
		return "", err
	}

	return r.ID, nil
}
