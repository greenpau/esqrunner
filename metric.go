package esqrunner

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

var supportedOperations map[string]bool
var supportedIndexSplit map[string]bool
var supportedFuctions map[string]bool

func init() {
	supportedOperations = make(map[string]bool)
	supportedIndexSplit = make(map[string]bool)
	supportedFuctions = make(map[string]bool)
	supportedOperations["GET"] = true
	supportedIndexSplit["daily"] = true
	supportedFuctions["_count"] = true
}

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
		return fmt.Errorf("attribute Name not set in %v", *m)
	}
	if m.Category == "" {
		return fmt.Errorf("attribute Category not set in %v", *m)
	}
	if m.Description == "" {
		return fmt.Errorf("attribute Description not set in %v", *m)
	}
	if m.Operation == "" {
		return fmt.Errorf("attribute Operation not set in %v", *m)
	}
	if m.BaseIndex == "" {
		return fmt.Errorf("attribute BaseIndex not set in %v", *m)
	}
	if m.IndexSplit == "" {
		return fmt.Errorf("attribute IndexSplit not set in %v", *m)
	}
	if m.Function == "" {
		return fmt.Errorf("attribute Function not set in %v", *m)
	}

	if _, supported := supportedOperations[m.Operation]; !supported {
		return fmt.Errorf(
			"attribute Operation has unsupported value: %s, metric: %v",
			m.Operation, *m,
		)
	}

	if _, supported := supportedIndexSplit[m.IndexSplit]; !supported {
		return fmt.Errorf(
			"attribute IndexSplit has unsupported value: %s, metric: %v",
			m.IndexSplit, *m,
		)
	}

	if _, supported := supportedFuctions[m.Function]; !supported {
		return fmt.Errorf(
			"attribute Function has unsupported value: %s, metric: %v",
			m.Function, *m,
		)
	}

	return nil
}
