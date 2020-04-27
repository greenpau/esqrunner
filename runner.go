package esqrunner

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

// QueryRunner is Elasticsearch query runner.
type QueryRunner struct {
	Config       *RunnerConfig
	Metrics      []*Metric
	ValidateOnly bool
}

// New return an instance of QueryRunner.
func New() *QueryRunner {
	return &QueryRunner{}
}

// ReadInConfig configures QueryRunner based on the provided
// configuration file.
func (r *QueryRunner) ReadInConfig(configFile string) error {
	log.Debugf("configuration file: %s", configFile)
	configDir, configFile := filepath.Split(configFile)
	ext := filepath.Ext(configFile)
	log.Debugf("configuration file extension: %s", ext)
	confSyntax := "yaml"
	switch ext {
	case ".yaml":
		log.Debugf("configuration syntax is yaml")
	case ".yml":
		log.Debugf("configuration syntax is yaml")
	default:
		return fmt.Errorf("configuration file type is unsupported")
	}
	log.Debugf("configuration directory: %s", configDir)
	log.Debugf("configuration syntax: %s", confSyntax)
	content, err := ioutil.ReadFile(filepath.Join(configDir, configFile))
	if err != nil {
		return err
	}
	config := RunnerConfig{}
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return err
	}
	r.Config = &config
	return nil
}

// ValidateConfig validates QueryRunner configuration.
func (r *QueryRunner) ValidateConfig() error {
	if r.Config == nil {
		return fmt.Errorf("configuration not found")
	}
	if err := r.Config.Validate(); err != nil {
		return err
	}
	return nil
}

// Run triggers the execution of the queries.
func (r *QueryRunner) Run() error {
	if err := r.ValidateConfig(); err != nil {
		return err
	}
	return nil
}
