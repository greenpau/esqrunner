package esqrunner

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// RunnerConfig is the configuration of the QueryRunner.
type RunnerConfig struct {
	MetricRef     map[string]*Metric   `json:"-" yaml:"-"`
	Metrics       []*Metric            `json:"-" yaml:"-"`
	Timestamps    []time.Time          `json:"-" yaml:"-"`
	MetricSources []string             `json:"metric_sources" yaml:"metric_sources"`
	Elasticsearch *ElasticsearchConfig `json:"elasticsearch" yaml:"elasticsearch"`
}

// Validate validates QueryRunner configuration.
func (c *RunnerConfig) Validate() error {
	if c.MetricRef == nil {
		c.MetricRef = make(map[string]*Metric)
	}
	if c.Elasticsearch == nil {
		return fmt.Errorf("no Elasticsearch configuration found")
	}
	if err := c.Elasticsearch.ValidateConfig(); err != nil {
		return err
	}
	if len(c.MetricSources) == 0 {
		return fmt.Errorf("no metric configuration files found")
	}
	for _, confFile := range c.MetricSources {
		metrics, err := NewMetricsFromFile(confFile)
		if err != nil {
			return fmt.Errorf("metric source %s parsing failed: %s", confFile, err)
		}
		if len(metrics) == 0 {
			return fmt.Errorf("metric source %s has no metrics", confFile)
		}
		for _, metric := range metrics {
			if err := metric.Valid(); err != nil {
				return fmt.Errorf(
					"metric source %s has invalid metric: %v, error: %s",
					confFile, metric, err,
				)
			}
			if _, exists := c.MetricRef[metric.ID]; exists {
				return fmt.Errorf(
					"metric source %s has invalid metric: %v, duplicate ID",
					confFile, metric,
				)
			}
			c.Metrics = append(c.Metrics, metric)
			c.MetricRef[metric.ID] = metric
		}
	}
	return nil
}

// AddDates adds past dates.
func (c *RunnerConfig) AddDates(s string) error {
	lastDaysRegex, err := regexp.Compile(`^last (\d+) days, interval (\d+) day`)
	if err != nil {
		return err
	}
	if m := lastDaysRegex.FindStringSubmatch(s); len(m) > 0 {
		days, _ := strconv.Atoi(m[1])
		interval, _ := strconv.Atoi(m[2])
		if days != 0 && interval != 0 {
			for i := days - 1; i >= 0; i-- {
				t := time.Now().UTC().Add(time.Duration(-1*i*interval*24) * time.Hour)
				c.Timestamps = append(c.Timestamps, t)
			}
		}
		return nil
	}

	//t := time.Now().UTC().Add(time.Duration(-1*hours) * time.Hour)
	//dt := fmt.Sprintf("%d%02d%02d", t.Year(), t.Month(), t.Day())
	//c.Dates = append(c.Dates, dt)

	return fmt.Errorf("unsupported dates pattern: %s", s)
}
