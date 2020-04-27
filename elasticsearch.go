package esqrunner

import (
	"fmt"
)

// EsConfig represents conntenction settings associated with Elasticsearch
// instance.
type EsConfig struct {
	Address string `json:"addr" yaml:"addr"`
}

// ValidateConfig validates EsConfig.
func (es *EsConfig) ValidateConfig() error {
	if es.Address == "" {
		return fmt.Errorf("Elasticsearch config has no address")
	}
	return nil
}
