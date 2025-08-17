package main

import (
	"bitbucket-metrics/bitbucket"
	"bitbucket-metrics/config"
	"bitbucket-metrics/metrics"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

func getEnvOrPanic(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Panicf("'%s' environment variable is mandatory", name)
	}
	return value
}

func getEnvOrDefault[T any](name string, defaultValue T) T {
	valueString := os.Getenv(name)
	if valueString == "" {
		log.Debugf("Variable '%s' is empty, returning default value '%v'", name, defaultValue)
		return defaultValue
	}

	var result any
	switch any(defaultValue).(type) {
	case string:
		result = valueString
	case int:
		number, err := strconv.Atoi(valueString)
		if err != nil {
			log.Warnf("Variable '%s' is not a valid int, returning default value '%v'", name, defaultValue)
			return defaultValue
		}
		result = number
	}

	return result.(T)
}

func parseLogLevel(level string) log.Level {
	var logLevels = map[string]log.Level{
		"debug":   log.DebugLevel,
		"info":    log.InfoLevel,
		"warn":    log.WarnLevel,
		"warning": log.WarnLevel,
		"error":   log.ErrorLevel,
		"fatal":   log.FatalLevel,
		"panic":   log.PanicLevel,
	}
	if lvl, ok := logLevels[level]; ok {
		return lvl
	}
	return log.InfoLevel
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	logLevel := strings.ToLower(getEnvOrDefault("LOG_LEVEL", "info"))
	log.SetLevel(parseLogLevel(logLevel))
	log.Info("Application started")

	configFilename := getEnvOrDefault("CONFIG", "config.yaml")
	config, err := config.ReadConfig(configFilename)
	if err != nil {
		log.WithFields(log.Fields{
			"filename": configFilename,
		}).Panic("Cannot load config file")
	}

	bitbucketBaseURL := getEnvOrPanic("BASE_URL")
	username := getEnvOrPanic("USERNAME")
	password := getEnvOrPanic("PASSWORD")
	apiPageSize := config.Bitbucket.ApiPageSize
	bitbucketRequestManager := bitbucket.Init(bitbucketBaseURL, username, password, apiPageSize)

	hostname := config.Bitbucket.Metrics.Hostname
	metricsPortNumber := uint16(config.Bitbucket.Metrics.Port)
	metricsPath := config.Bitbucket.Metrics.Path
	metrics.ListenAndServe(hostname, metricsPortNumber, metricsPath, func(metricsToBeCollected *metrics.Metrics) {
		bitbucket.NewRunner(config, bitbucketRequestManager, metricsToBeCollected)
	})

	log.Info("Application stopped")
}
