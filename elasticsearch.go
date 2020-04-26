package main

// EsConfig represents conntenction settings associated with Elasticsearch
// instance.
type EsConfig struct {
	Address string `json:"addr" yaml:"addr"`
}
