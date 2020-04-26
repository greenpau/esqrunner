package main

import (
	"encoding/json"
)

// MetricConfig is a collection of attrbutes and parameters
// for the creation of a metric.
type MetricConfig struct {
	Category    string           `json:"category" yaml:"category"`
	Name        string           `json:"name" yaml:"name"`
	Description string           `json:"description" yaml:"description"`
	Operation   string           `json:"operation" yaml:"operation"`
	BaseIndex   string           `json:"base_index" yaml:"base_index"`
	IndexSplit  string           `json:"index_split" yaml:"index_split"`
	Function    string           `json:"dsl_function" yaml:"dsl_function"`
	Query       *json.RawMessage `json:"dsl_query" yaml:"dsl_query"`
}
