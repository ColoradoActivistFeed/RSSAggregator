package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"time"

	"ActivistFeed/aggregator"

	log "github.com/sirupsen/logrus"
)



func main() {

	configPath := flag.String("config", "config.json", "")
	outputSample := flag.Bool("sample", false, "")
	loadFile := flag.String("file", "", "")
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
			Name:        "",
			Description: "",
			Organizations: map[string]aggregator.Organization{
				"": {
					Description: "",
					Sources:     []string{},
					Author:      "",
					Link:        "",
					Slug:        "",
				},
			},
			RefreshIntervalSeconds: 0,
			TemplatePath:           "",
			OutputPath:             "",
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

	if *loadFile != "" {
		agg, err := aggregator.NewAggregator(*configPath)
		if err != nil {
			log.WithError(err).Fatalf("failed to initialize aggregator")
		} else {

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
		}
	} else {
		agg, err := aggregator.NewAggregator(*configPath)
		if err != nil {
			log.WithError(err).Fatalf("failed to initialize aggregator")
		} else {
			if err := agg.Run(); err != nil {
				//if err := agg.Start(); err != nil {
				log.WithError(err).Fatalf("failed to run aggregator")
			}
		}

		data, err := json.MarshalIndent(agg.Content, "", "  ")
		if err != nil {
			log.WithError(err).Fatalf("failed convert data to json")
		}
		if err := ioutil.WriteFile("data.json", data, 0644); err != nil {
			log.WithError(err).Fatalf("failed save json data")
		}
	}
	}
