package esqrunner

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	//"github.com/elastic/go-elasticsearch/v7/esapi"
	"bytes"
	"context"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"time"
)

// ElasticsearchConfig represents conntenction settings associated with Elasticsearch
// instance.
type ElasticsearchConfig struct {
	Address []string `json:"addr" yaml:"addr"`
}

// ValidateConfig validates ElasticsearchConfig.
func (es *ElasticsearchConfig) ValidateConfig() error {
	if len(es.Address) < 1 {
		return fmt.Errorf("Elasticsearch config has no address")
	}
	return nil
}

// ElasticsearchClient is Elasticsearch client.
type ElasticsearchClient struct {
	driver *elasticsearch7.Client
}

// NewElasticsearchClient returns Elasticsearch instance.
func NewElasticsearchClient(cfg *ElasticsearchConfig) (*ElasticsearchClient, error) {
	c := &ElasticsearchClient{}

	esConfig := elasticsearch7.Config{
		Addresses: cfg.Address,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Duration(5) * time.Second,
			DialContext:           (&net.Dialer{Timeout: time.Duration(5) * time.Second}).DialContext,
			TLSClientConfig: &tls.Config{
				MaxVersion:         tls.VersionTLS13,
				InsecureSkipVerify: true,
			},
		},
	}

	client, err := elasticsearch7.NewClient(esConfig)
	if err != nil {
		return nil, err
	}
	c.driver = client

	return c, nil
}

// ElasticsearchInfo contains server info.
type ElasticsearchInfo struct {
	Version string
}

// Info returns Elasticsearch info.
func (c *ElasticsearchClient) Info() (*ElasticsearchInfo, error) {
	res, err := c.driver.Info()
	if err != nil {
		return nil, fmt.Errorf("elasticsearch connection error: %s", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch query error: %s", res.String())
	}
	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("Error parsing elasticsearch response body: %s", err)
	}
	info := &ElasticsearchInfo{}
	info.Version = r["version"].(map[string]interface{})["number"].(string)
	return info, nil
}

// ElasticsearchCounter is a counter.
type ElasticsearchCounter struct {
	Total uint64
}

// Count returns total counter frim Elasticsearch query
//
// References:
//
// - [Elasticsearch Reference - Count API](https://www.elastic.co/guide/en/elasticsearch/reference/master/search-count.html)
//
// - [package esapi](https://pkg.go.dev/github.com/elastic/go-elasticsearch/v7@v7.6.0/esapi?tab=doc#Count)
func (c *ElasticsearchClient) Count(m *Metric, suffix string) (*ElasticsearchCounter, error) {
	counter := &ElasticsearchCounter{}
	if m.Function != "_count" || m.Operation != "GET" {
		return nil, fmt.Errorf("metric %v does not support Count()", *m)
	}
	index := fmt.Sprintf("%s%s", m.BaseIndex, suffix)
	qbuf := []byte(*m.Query)
	buf := bytes.NewBuffer(qbuf)
	res, err := c.driver.Count(
		c.driver.Count.WithContext(context.Background()),
		c.driver.Count.WithIndex(index),
		c.driver.Count.WithBody(buf),
		c.driver.Count.WithPretty(),
	)
	/*
		This is search
		res, err := c.driver.Search(
			c.driver.Search.WithContext(context.Background()),
			c.driver.Search.WithIndex(index),
			c.driver.Search.WithBody(buf),
			c.driver.Search.WithTrackTotalHits(true),
			c.driver.Search.WithPretty(),
		)
	*/
	if err != nil {
		return nil, fmt.Errorf("elasticsearch connection error: %s, metric: %v", err, *m)
	}
	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch query error: %s, metric: %v", res.String(), *m)
	}

	log.Debugf("elasticsearch responded with: %s", res.String())

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("Error parsing elasticsearch response body: %s, metric: %v", err, *m)
	}

	if _, ok := r["count"]; !ok {
		return nil, fmt.Errorf("No total in elasticsearch response body: %v, metric: %v", r, *m)
	}
	count := r["count"].(float64)
	counter.Total = uint64(count)
	return counter, nil
}
