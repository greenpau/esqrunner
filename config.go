package esqrunner

import "fmt"

// RunnerConfig is the configuration of the QueryRunner.
type RunnerConfig struct {
	MetricSources []string  `json:"metric_sources" yaml:"metric_sources"`
	Elasticsearch *EsConfig `json:"elasticsearch" yaml:"elasticsearch"`
}

// Validate validates QueryRunner configuration.
func (c *RunnerConfig) Validate() error {
	if c.Elasticsearch == nil {
		return fmt.Errorf("no Elasticsearch configuration found")
	}
	if err := c.Elasticsearch.ValidateConfig(); err != nil {
		return err
	}
	if len(c.MetricSources) == 0 {
		return fmt.Errorf("no metric configuration files found")
	}
	for _, metricConfigFile := range c.MetricSources {
		metrics, err := NewMetricsFromFile(metricConfigFile)
		if err != nil {
			return fmt.Errorf("metric source %s parsing failed: %s", metricConfigFile, err)
		}
		if len(metrics) == 0 {
			return fmt.Errorf("metric source %s has no metrics", metricConfigFile)
		}
	}
	return nil
}
