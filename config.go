package esqrunner

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"regexp"
	"sort"
	"strconv"
	"time"
)

// RunnerConfig is the configuration of the QueryRunner.
type RunnerConfig struct {
	MetricRef  map[string]*Metric `json:"-" yaml:"-"`
	Metrics    []*Metric          `json:"-" yaml:"-"`
	Timestamps []time.Time        `json:"-" yaml:"-"`
	Output     struct {
		Landscape bool   `json:"-" yaml:"-"`
		Format    string `json:"-" yaml:"-"`
	}
	MetricSources []string             `json:"metric_sources" yaml:"metric_sources"`
	Elasticsearch *ElasticsearchConfig `json:"elasticsearch" yaml:"elasticsearch"`
	Metadata      struct {
		FieldList []string       `json:"-" yaml:"-"`
		Fields    map[string]int `json:"-" yaml:"-"`
		Size      int            `json:"-" yaml:"-"`
	}
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

	supportedFormats := map[string]bool{
		"csv":  true,
		"json": true,
	}
	if c.Output.Format == "" {
		c.Output.Format = "csv"
	}

	if _, exists := supportedFormats[c.Output.Format]; !exists {
		return fmt.Errorf("the following output format is not supported: %s", c.Output.Format)
	}

	log.Debugf("output format: %s", c.Output.Format)

	if len(c.MetricSources) == 0 {
		return fmt.Errorf("no metric configuration files found")
	}

	if c.Metadata.Fields == nil {
		c.Metadata.Fields = make(map[string]int)
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
			if len(metric.Metadata) > 0 {
				for k := range metric.Metadata {
					if _, exists := c.Metadata.Fields[k]; !exists {
						c.Metadata.Fields[k] = 1
					} else {
						c.Metadata.Fields[k]++
					}
				}
				if len(c.Metadata.Fields) > c.Metadata.Size {
					c.Metadata.Size = len(c.Metadata.Fields)
				}
			}
			c.Metrics = append(c.Metrics, metric)
			c.MetricRef[metric.ID] = metric
		}
	}

	if c.Metadata.Size > 0 {
		for k, v := range c.Metadata.Fields {
			c.Metadata.FieldList = append(c.Metadata.FieldList, k)
			log.Debugf("configuration metadata field size: %s = %d", k, v)
		}
		sort.Strings(c.Metadata.FieldList)
		for i, v := range c.Metadata.FieldList {
			c.Metadata.FieldList[i] = v
		}
		log.Debugf("configuration metadata fields: %v", c.Metadata.FieldList)
	}
	log.Debugf("configuration metadata width: %d", c.Metadata.Size)
	log.Debugf("total metrics: %d", len(c.Metrics))
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
				//t := time.Now().UTC().Add(time.Duration(-1*i*interval*24) * time.Hour)
				t := time.Now().Add(time.Duration(-1*i*interval*24) * time.Hour)
				c.Timestamps = append(c.Timestamps, t)
			}
		}
		return nil
	}

	return fmt.Errorf("unsupported dates pattern: %s", s)
}
