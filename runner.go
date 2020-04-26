package main

// QueryRunner is Elasticsearch query runner.
type QueryRunner struct {
	Config       *RunnerConfig
	ValidateOnly bool
}

// RunnerConfig is the configuration of the QueryRunner.
type RunnerConfig struct {
	MetricSources []string  `json:"metric_sources" yaml:"metric_sources"`
	Elasticsearch *EsConfig `json:"elasticsearch" yaml:"elasticsearch"`
}

// New return an instance of QueryRunner.
func New() *QueryRunner {
	return &QueryRunner{}
}

// Run triggers the execution of the queries.
func (r *QueryRunner) Run() error {
	return nil
}
