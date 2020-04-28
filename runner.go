package esqrunner

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"path/filepath"
	"strings"
)

// QueryRunner is Elasticsearch query runner.
type QueryRunner struct {
	client       *ElasticsearchClient
	Config       *RunnerConfig
	Metrics      map[string][]uint64
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
	content, err := readFileBytes(filepath.Join(configDir, configFile))
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

	if r.Metrics == nil {
		r.Metrics = make(map[string][]uint64)
	}

	client, err := NewElasticsearchClient(r.Config.Elasticsearch)
	if err != nil {
		return err
	}

	r.client = client
	srv, err := r.client.Info()
	if err != nil {
		return err
	}
	log.Debugf("Elasticsearch server version: %s", srv.Version)

	for _, ts := range r.Config.Timestamps {
		log.Debugf("Processing date: %s", ts)
		for _, m := range r.Config.Metrics {
			if m.Disabled {
				continue
			}
			if m.IndexSplit != "daily" {
				continue
			}
			if _, exists := r.Metrics[m.ID]; !exists {
				r.Metrics[m.ID] = []uint64{}
			}
			suffix := fmt.Sprintf("%d%02d%02d", ts.Year(), ts.Month(), ts.Day())
			count, err := r.client.Count(m, suffix)
			if err != nil {
				log.Errorf("Elasticsearch query error: %s", err)
				return err
			}
			r.Metrics[m.ID] = append(r.Metrics[m.ID], count.Total)
		}
	}

	var sb strings.Builder
	line := []string{}
	line = append(line, "Metrics")
	for _, ts := range r.Config.Timestamps {
		line = append(line, ts.Format("2006/01/02"))
	}
	line = append(line, "Metric ID")
	sb.WriteString(strings.Join(line, "|") + "\n")

	for _, m := range r.Config.Metrics {
		if m.Disabled {
			continue
		}
		line = []string{}
		line = append(line, m.Name)
		for _, count := range r.Metrics[m.ID] {
			line = append(line, fmt.Sprintf("%d", count))
		}
		line = append(line, m.ID)
		sb.WriteString(strings.Join(line, "|") + "\n")
	}
	fmt.Println(sb.String())

	return nil
}
