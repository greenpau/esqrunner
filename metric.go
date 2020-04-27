package esqrunner

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

// Metric is a collection of attrbutes and parameters
// for the creation and management of a metric.
type Metric struct {
	Category    string           `json:"category" yaml:"category"`
	Name        string           `json:"name" yaml:"name"`
	Description string           `json:"description" yaml:"description"`
	Operation   string           `json:"operation" yaml:"operation"`
	BaseIndex   string           `json:"base_index" yaml:"base_index"`
	IndexSplit  string           `json:"index_split" yaml:"index_split"`
	Function    string           `json:"dsl_function" yaml:"dsl_function"`
	Query       *json.RawMessage `json:"dsl_query" yaml:"dsl_query"`
}

// NewMetricsFromFile parses a JSON file containing metrics, and
// return a collection of Metric instances.
func NewMetricsFromFile(configFile string) ([]*Metric, error) {
	log.Debugf("metric configuration file: %s", configFile)
	configDir, configFile := filepath.Split(configFile)
	ext := filepath.Ext(configFile)
	confSyntax := "json"
	switch ext {
	case ".json":
		log.Debugf("metric configuration syntax is json")
	default:
		return []*Metric{}, fmt.Errorf("configuration file type is unsupported")
	}
	if confSyntax != "json" {
		return []*Metric{}, fmt.Errorf("configuration file syntax is unsupported: %s", confSyntax)
	}
	content, err := readFileBytes(filepath.Join(configDir, configFile))
	if err != nil {
		return []*Metric{}, err
	}
	metrics := []*Metric{}
	err = json.Unmarshal(content, &metrics)
	if err != nil {
		return []*Metric{}, err
	}

	return metrics, nil
}

// Valid validates whether a metric definition has mandatory fields and
// that the fields conform to a standard set in this function.
func (m *Metric) Valid() error {
	if m.Name == "" {
		return fmt.Errorf("attribute Name not found")
	}
	return nil
}
