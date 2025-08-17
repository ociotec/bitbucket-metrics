package config

import (
	"os"
	"strconv"
	"testing"
)

func TestReadConfigWithNotExistingConfigFile(t *testing.T) {
	config, err := ReadConfig("notexistingfile.yaml")
	if config != nil {
		t.Fatal("Unexpect not empty config with an invalid config file")
	}
	if err == nil {
		t.Fatal("Unexpected empty error with an invalid config file")
	}
}

func createTempConfig(t *testing.T, content string) string {
	t.Helper()

	f, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatalf("Cannot create temp config file %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("Cannot write temp config file %v", f.Name())
	}

	return f.Name()
}

func TestReadConfigWithInvalidYamlSyntax(t *testing.T) {
	filename := createTempConfig(t, "Invalid YAML syntax")
	defer os.Remove(filename)

	config, err := ReadConfig(filename)
	if config != nil {
		t.Fatal("Unexpect not empty config with an invalid config file")
	}
	if err == nil {
		t.Fatal("Unexpected empty error with an invalid config file")
	}
}

const EXPECTED_API_PAGE_SIZE = 123
const EXPECTED_HOSTNAME = "hostname"
const EXPECTED_PORT = 1234
const EXPECTED_PATH = "/expected/path"
const EXPECTED_PERIOD_IN_SECONDS = 12

var EXPECTED_PROJECTS_INCLUDE = []string{"project1", "project2"}

var CONFIG_CONTENT = "bitbucket:\n" +
	"  api_page_size: " + strconv.Itoa(EXPECTED_API_PAGE_SIZE) + "\n" +
	"  metrics:\n" +
	"    hostname: " + EXPECTED_HOSTNAME + "\n" +
	"    port: " + strconv.Itoa(EXPECTED_PORT) + "\n" +
	"    path: " + EXPECTED_PATH + "\n" +
	"    period_in_seconds: " + strconv.Itoa(EXPECTED_PERIOD_IN_SECONDS) + "\n" +
	"  projects:\n" +
	"    include:\n" +
	"      - " + EXPECTED_PROJECTS_INCLUDE[0] + "\n" +
	"      - " + EXPECTED_PROJECTS_INCLUDE[1] + "\n"

func TestReadConfig(t *testing.T) {
	filename := createTempConfig(t, CONFIG_CONTENT)
	defer os.Remove(filename)

	config, err := ReadConfig(filename)
	if err != nil {
		t.Fatalf("Fail to read testing config %v", filename)
	}
	if config.Bitbucket.ApiPageSize != EXPECTED_API_PAGE_SIZE {
		t.Errorf("bitbucket.api_page_size should be %v instead of %v", EXPECTED_API_PAGE_SIZE, config.Bitbucket.ApiPageSize)
	}
	if config.Bitbucket.Metrics.Hostname != EXPECTED_HOSTNAME {
		t.Errorf("bitbucket.metrics.hostname should be %v instead of %v", EXPECTED_HOSTNAME, config.Bitbucket.Metrics.Hostname)
	}
	if config.Bitbucket.Metrics.Port != EXPECTED_PORT {
		t.Errorf("bitbucket.metrics.port should be %v instead of %v", EXPECTED_PORT, config.Bitbucket.Metrics.Port)
	}
	if config.Bitbucket.Metrics.Path != EXPECTED_PATH {
		t.Errorf("bitbucket.metrics.path should be %v instead of %v", EXPECTED_PATH, config.Bitbucket.Metrics.Path)
	}
	if config.Bitbucket.Metrics.PeriodInSeconds != EXPECTED_PERIOD_IN_SECONDS {
		t.Errorf("bitbucket.metrics.period_in_seconds should be %v instead of %v", EXPECTED_PERIOD_IN_SECONDS, config.Bitbucket.Metrics.PeriodInSeconds)
	}
	if len(config.Bitbucket.Projects.Include) != len(EXPECTED_PROJECTS_INCLUDE) {
		t.Errorf("bitbucket.projects.include length should be %v instead of %v", len(EXPECTED_PROJECTS_INCLUDE), len(config.Bitbucket.Projects.Include))
		return
	}
	for i, project := range config.Bitbucket.Projects.Include {
		if project != EXPECTED_PROJECTS_INCLUDE[i] {
			t.Errorf("bitbucket.projects.include[%v] should be %v instead of %v", i, EXPECTED_PROJECTS_INCLUDE[i], project)
		}
	}
}
