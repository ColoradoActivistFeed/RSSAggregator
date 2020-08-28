package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"time"

	"github.com/ColoradoActivistFeed/RSSAggregator/aggregator"

	log "github.com/sirupsen/logrus"
)

func main() {

	configPath := flag.String("config", "config.json", "")
	outputSample := flag.Bool("sample", false, "")
	loadFile := flag.String("load", "", "")
	saveFile := flag.String("save", "", "")
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{
		ForceColors:               false,
		EnvironmentOverrideColors: true,
		DisableTimestamp:          false,
		FullTimestamp:             true,
		TimestampFormat:           time.RFC3339,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	if *outputSample {
		cfg := &aggregator.Config{
			Name:                   "",
			Description:            "",
			Link:                   "",
			AutoCommit:             true,
			RefreshIntervalSeconds: 0,
			TemplatePath:           "templates",
			OutputPath:             "docs",
			StaticAssets:           "static",
			TimeZone:               "America/Denver",
			Organizations: map[string]aggregator.Organization{
				"": {
					Description: "",
					Author:      "",
					Slug:        "",
					Link:        "",
					Sources:     []string{},
				},
			},
		}
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			log.WithError(err).Fatalf("could not marshal json")
		}
		if err := ioutil.WriteFile(*configPath, data, 0644); err != nil {
			log.WithError(err).Fatalf("could not write config")
		}
		os.Exit(0)
	}

	agg, err := aggregator.NewAggregator(*configPath)
	if err != nil {
		log.WithError(err).Fatalf("failed to initialize aggregator")
	}

	if *loadFile != "" {
		data, err := ioutil.ReadFile(*loadFile)
		if err != nil {
			log.WithError(err).Fatalf("failed read saved data")
		}

		if err := json.Unmarshal(data, &agg.Content); err != nil {
			log.WithError(err).Fatalf("failed read saved data")
		}

		if err := agg.WriteRSS(); err != nil {
			log.WithError(err).Fatalf("failed to run aggregator")
		}
		if err := agg.WriteHTML(); err != nil {
			log.WithError(err).Fatalf("failed to run aggregator")
		}
	} else if *saveFile != "" {
		if err := agg.Run(); err != nil {
			log.WithError(err).Fatalf("failed to run aggregator")
		}

		data, err := json.MarshalIndent(agg.Content, "", "  ")
		if err != nil {
			log.WithError(err).Fatalf("failed convert data to json")
		}
		if err := ioutil.WriteFile("data.json", data, 0644); err != nil {
			log.WithError(err).Fatalf("failed save json data")
		}
	} else {
		if err := agg.Start(); err != nil {
			log.WithError(err).Fatalf("failed to run aggregator")
		}
	}
}
