package config

import (
	"os"

	"gopkg.in/yaml.v3"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	Bitbucket Bitbucket `yaml:"bitbucket"`
}

type Bitbucket struct {
	ApiPageSize int      `yaml:"api_page_size"`
	Metrics     Metrics  `yaml:"metrics"`
	Projects    Projects `yaml:"projects"`
}

type Metrics struct {
	Hostname        string `yaml:"hostname"`
	Port            int    `yaml:"port"`
	Path            string `yaml:"path"`
	PeriodInSeconds int    `yaml:"period_in_seconds"`
}

type Projects struct {
	Include []string `yaml:"include"`
}

func ReadConfig(filename string) (*Config, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		log.WithFields(log.Fields{
			"filename": filename,
			"error":    err,
		}).Error("Cannot read config file")
		return nil, err
	}

	config := Config{
		Bitbucket: Bitbucket{
			ApiPageSize: 100,
			Metrics: Metrics{
				Hostname:        "localhost",
				Port:            8080,
				Path:            "/metrics",
				PeriodInSeconds: 600,
			},
			Projects: Projects{
				Include: nil,
			},
		},
	}
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		log.WithFields(log.Fields{
			"filename": filename,
			"error":    err,
		}).Error("Cannot parse YAML content from config file")
		return nil, err
	}

	log.WithFields(log.Fields{
		"filename": filename,
	}).Info("YAML content parsed from config file")

	return &config, nil
}
