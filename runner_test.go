// Copyright 2020 Paul Greenberg (greenpau@outlook.com)

package esqrunner

import (
	"testing"
)

func TestRunner(t *testing.T) {
	confFile := "assets/conf/default.yaml"
	r := New()

	if err := r.ReadInConfig(confFile); err != nil {
		t.Fatalf("error loading config %s: %s", confFile, err)
	}

	if err := r.ValidateConfig(); err != nil {
		t.Fatalf("error validating config: %s", err)
	}

	expConfigMetrics := 1
	if len(r.Config.Metrics) != expConfigMetrics {
		t.Fatalf(
			"error validating config, expected %d metrics, received: %d",
			expConfigMetrics, len(r.Config.Metrics),
		)
	}

	t.Logf("Elasticsearch address: %s", r.Config.Elasticsearch.Address)
}
