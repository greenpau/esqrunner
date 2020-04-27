package esqrunner

import "fmt"

// RunnerConfig is the configuration of the QueryRunner.
type RunnerConfig struct {
	MetricSources []string  `json:"metric_sources" yaml:"metric_sources"`
	Elasticsearch *EsConfig `json:"elasticsearch" yaml:"elasticsearch"`
}

// ValidateConfig validates QueryRunner configuration.
func (r *QueryRunner) ValidateConfig() error {
	if r.Config == nil {
		return fmt.Errorf("Configuration not found")
	}
	if r.Config.Elasticsearch == nil {
		return fmt.Errorf("No Elasticsearch configuration found")
	}
	if err := r.Config.Elasticsearch.ValidateConfig(); err != nil {
		return err
	}
	return nil
}
