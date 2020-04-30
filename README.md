# Elasticsearch Query Runner

<a href="https://github.com/greenpau/esqrunner/actions/" target="_blank"><img src="https://github.com/greenpau/esqrunner/workflows/build/badge.svg?branch=master"></a>
<a href="https://pkg.go.dev/github.com/greenpau/esqrunner" target="_blank"><img src="https://img.shields.io/badge/godoc-reference-blue.svg"></a>

Run Elasticsearh queries and create metrics based on the result
of the queries in Elasticsearch database.

## Getting Started

First, define the metrics to collect from Elasticsearch. Please see
`assets/metrics/simple.json` for an example.

```json
[
  {
    "id": "28e3c0fb594443fea16131c5f26eeb81",
    "category": "Helpdesk",
    "name": "Helpdesk Ticket Total",
    "description": "The total number of helpdesk tickets",
    "operation": "GET",
    "base_index": "tickets-",
    "index_split": "daily",
    "dsl_function": "_count",
    "dsl_query": {
      "query": {
        "bool": {
          "must_not": [
            {
              "match_phrase": {
                "classification": "internal"
              }
            }
          ]
        }
      }
    }
  }
]
```

Next, create a configuration file pointing to the metrics and
identifying Elasticsearch:

```yaml
---
metric_sources:
  - 'assets/metrics/simple.json'
elasticsearch:
  addr:
    - 'http://localhost:9200'
```

Finally, run `esqrunner` tool to create datasets:

```bash
./bin/esqrunner --config ~/tmpelastic/config.yaml --log-level debug --datepicker "last 7 days, interval 1 day" --output-file-prefix "metrics_last_7d"
```

The expected output follows:

```
Output file prefix: /tmp/esqrunner-703462495/metrics_last_7d
Wrote data to /tmp/esqrunner-703462495/metrics_last_7d_landscape.csv
Wrote data to /tmp/esqrunner-703462495/metrics_last_7d_portrait.csv
Wrote data to /tmp/esqrunner-703462495/metrics_last_7d.json
Wrote data to /tmp/esqrunner-703462495/metrics_last_7d.js
```
