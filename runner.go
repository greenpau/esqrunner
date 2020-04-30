package esqrunner

import (
	"encoding/json"
	"fmt"
	"github.com/greenpau/go-calculator"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"strings"
)

// QueryRunner is Elasticsearch query runner.
type QueryRunner struct {
	client       *ElasticsearchClient
	Config       *RunnerConfig
	Metrics      map[string][]uint64
	MetricErrors map[string][]error
	ValidateOnly bool
}

// New return an instance of QueryRunner.
func New() *QueryRunner {
	return &QueryRunner{}
}

// ReadInConfig configures QueryRunner based on the provided
// configuration file.
func (r *QueryRunner) ReadInConfig(configFile string) error {
	log.Debugf("configuration file: %s", configFile)
	configDir, configFile := filepath.Split(configFile)
	ext := filepath.Ext(configFile)
	log.Debugf("configuration file extension: %s", ext)
	confSyntax := "yaml"
	switch ext {
	case ".yaml":
		log.Debugf("configuration syntax is yaml")
	case ".yml":
		log.Debugf("configuration syntax is yaml")
	default:
		return fmt.Errorf("configuration file type is unsupported")
	}
	log.Debugf("configuration directory: %s", configDir)
	log.Debugf("configuration syntax: %s", confSyntax)
	content, err := readFileBytes(filepath.Join(configDir, configFile))
	if err != nil {
		return err
	}
	config := RunnerConfig{}
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return err
	}
	r.Config = &config
	return nil
}

// ValidateConfig validates QueryRunner configuration.
func (r *QueryRunner) ValidateConfig() error {
	if r.Config == nil {
		return fmt.Errorf("configuration not found")
	}
	if err := r.Config.Validate(); err != nil {
		return err
	}
	return nil
}

// Run triggers the execution of the queries.
func (r *QueryRunner) Run() error {
	if err := r.ValidateConfig(); err != nil {
		return err
	}

	if r.Metrics == nil {
		r.Metrics = make(map[string][]uint64)
	}

	if r.MetricErrors == nil {
		r.MetricErrors = make(map[string][]error)
	}

	client, err := NewElasticsearchClient(r.Config.Elasticsearch)
	if err != nil {
		return err
	}

	r.client = client
	srv, err := r.client.Info()
	if err != nil {
		return err
	}
	log.Debugf("Elasticsearch server version: %s", srv.Version)

	for _, ts := range r.Config.Timestamps {
		log.Debugf("Processing date: %s", ts)
		for _, m := range r.Config.Metrics {
			if m.Disabled {
				continue
			}
			if m.IndexSplit != "daily" {
				continue
			}
			if _, exists := r.Metrics[m.ID]; !exists {
				r.Metrics[m.ID] = []uint64{}
			}
			if _, exists := r.MetricErrors[m.ID]; !exists {
				r.MetricErrors[m.ID] = []error{}
			}
			suffix := fmt.Sprintf("%d%02d%02d", ts.Year(), ts.Month(), ts.Day())
			count, err := r.client.Count(m, suffix)
			if err != nil {
				r.Metrics[m.ID] = append(r.Metrics[m.ID], 0)
				r.MetricErrors[m.ID] = append(r.MetricErrors[m.ID], err)
				continue
			}
			r.Metrics[m.ID] = append(r.Metrics[m.ID], count.Total)
			r.MetricErrors[m.ID] = append(r.MetricErrors[m.ID], nil)
		}
	}
	return nil
}

// Output returns metrics data.
func (r *QueryRunner) Output() (string, error) {
	var sb strings.Builder
	if r.Config.Output.Format == "csv" {
		sp := ";"
		if r.Config.Output.Landscape {
			line := []string{}
			line = append(line, "Categories")
			line = append(line, "Metrics")
			for _, k := range r.Config.Metadata.FieldList {
				line = append(line, strings.Title(k))
			}
			for _, ts := range r.Config.Timestamps {
				line = append(line, ts.Format("2006/01/02"))
			}
			line = append(line, "Total")
			line = append(line, "Max")
			line = append(line, "Min")
			line = append(line, "Average")
			line = append(line, "Median")
			line = append(line, "Modes")
			line = append(line, "Range")
			line = append(line, "Metric ID")
			sb.WriteString(strings.Join(line, sp) + "\n")

			for _, m := range r.Config.Metrics {
				if m.Disabled {
					continue
				}
				line = []string{}
				line = append(line, m.Category)
				line = append(line, m.Name)

				for _, k := range r.Config.Metadata.FieldList {
					if v, exists := m.Metadata[k]; exists {
						line = append(line, fmt.Sprintf("%s", v))
					} else {
						line = append(line, "-")
					}
				}

				for i, count := range r.Metrics[m.ID] {
					if r.MetricErrors[m.ID][i] == nil {
						line = append(line, fmt.Sprintf("%d", count))
					} else {
						line = append(line, "-")
					}
				}
				calc := calculator.NewUint64(r.Metrics[m.ID])
				calc.RunAll()
				line = append(line, fmt.Sprintf("%.2f", calc.Register.Total))
				line = append(line, fmt.Sprintf("%.2f", calc.Register.MaxValue))
				line = append(line, fmt.Sprintf("%.2f", calc.Register.MinValue))
				line = append(line, fmt.Sprintf("%.2f", calc.Register.Mean))
				line = append(line, fmt.Sprintf("%.2f", calc.Register.Median))
				line = append(line, fmt.Sprintf("%v", calc.Register.Modes))
				line = append(line, fmt.Sprintf("%.2f", calc.Register.Range))
				line = append(line, m.ID)
				sb.WriteString(strings.Join(line, sp) + "\n")
			}
		} else {
			line := []string{}
			line = append(line, "Date")
			line = append(line, "Value")
			line = append(line, "Category")
			line = append(line, "Metric Name")
			for _, k := range r.Config.Metadata.FieldList {
				line = append(line, strings.Title(k))
			}

			line = append(line, "Metric ID")
			sb.WriteString(strings.Join(line, sp) + "\n")
			for _, m := range r.Config.Metrics {
				if m.Disabled {
					continue
				}
				for i, ts := range r.Config.Timestamps {
					line := []string{}
					line = append(line, ts.Format("2006/01/02"))

					if r.MetricErrors[m.ID][i] == nil {
						line = append(line, fmt.Sprintf("%d", r.Metrics[m.ID][i]))
					} else {
						line = append(line, "-")
					}
					line = append(line, m.Category)
					line = append(line, m.Name)
					for _, k := range r.Config.Metadata.FieldList {
						if v, exists := m.Metadata[k]; exists {
							line = append(line, fmt.Sprintf("%s", v))
						} else {
							line = append(line, "-")
						}
					}

					line = append(line, m.ID)
					sb.WriteString(strings.Join(line, sp) + "\n")
				}
			}
		}
	}

	if r.Config.Output.Format == "json" || r.Config.Output.Format == "js" {
		if r.Config.Output.Format == "js" {
			sb.WriteString("var metricsDataset = ")
		}
		sb.WriteString("{\n")
		metricDefinitions, err := json.MarshalIndent(r.Config.Metrics, r.offset(1), r.offset(1))
		if err != nil {
			return "", err
		}
		sb.WriteString(r.offset(1) + `"metric_definitions": ` + string(metricDefinitions) + ",\n")
		metricTimestamps := []string{}
		for _, ts := range r.Config.Timestamps {
			metricTimestamps = append(metricTimestamps, fmt.Sprintf("%d000", ts.Unix()))
		}
		sb.WriteString(r.offset(1) + `"timestamps": [` + strings.Join(metricTimestamps, ", ") + "],\n")
		sb.WriteString(r.offset(1) + `"metrics": [` + "\n")
		for j, m := range r.Config.Metrics {
			isLastMetricElement := false
			if len(r.Config.Metrics)-1 == j {
				isLastMetricElement = true
			}
			if m.Disabled {
				continue
			}
			sb.WriteString(r.offset(2) + "{\n")
			sb.WriteString(r.offset(3) + fmt.Sprintf(`"%s": {`, m.ID) + "\n")
			sb.WriteString(r.offset(4) + `"counters": [`)
			for i := range r.Config.Timestamps {
				isLastElement := false
				if len(r.Config.Timestamps)-1 == i {
					isLastElement = true
				}
				if r.MetricErrors[m.ID][i] == nil {
					sb.WriteString(fmt.Sprintf("%d", r.Metrics[m.ID][i]))
				} else {
					sb.WriteString("null")
				}
				if !isLastElement {
					sb.WriteString(", ")
				}
			}
			sb.WriteString("],\n")
			calc := calculator.NewUint64(r.Metrics[m.ID])
			calc.RunAll()
			sb.WriteString(r.offset(4) + fmt.Sprintf(`"total": %.2f,`, calc.Register.Total) + "\n")
			sb.WriteString(r.offset(4) + fmt.Sprintf(`"max": %.2f,`, calc.Register.MaxValue) + "\n")
			sb.WriteString(r.offset(4) + fmt.Sprintf(`"min": %.2f,`, calc.Register.MinValue) + "\n")
			sb.WriteString(r.offset(4) + fmt.Sprintf(`"mean": %.2f,`, calc.Register.Mean) + "\n")
			sb.WriteString(r.offset(4) + fmt.Sprintf(`"median": %.2f,`, calc.Register.Median) + "\n")
			sb.WriteString(r.offset(4) + `"modes": [`)
			if len(calc.Register.Modes) == 1000000 {
				rml := len(calc.Register.Modes) - 1
				for ri, rm := range calc.Register.Modes {
					sb.WriteString(fmt.Sprintf("%.2f", rm))
					if rml != ri {
						sb.WriteString(", ")
					}
				}
			}
			sb.WriteString("],\n")
			sb.WriteString(r.offset(4) + fmt.Sprintf(`"range": %.2f`, calc.Register.Range) + "\n")

			sb.WriteString(r.offset(3) + "}\n")
			if !isLastMetricElement {
				sb.WriteString(r.offset(2) + "},\n")
			} else {
				sb.WriteString(r.offset(2) + "}\n")
			}
			//sb.WriteString("]\n")
		}

		//			sb.WriteString(r.offset(2) + fmt.Sprintf("%d", ts.Unix()) + "000,\n")

		//sb.WriteString(r.offset(2) + "]\n")
		sb.WriteString(r.offset(1) + "]\n")
		sb.WriteString("}\n")
	}

	return sb.String(), nil
}

func (r *QueryRunner) offset(j int) string {
	var offset string
	for i := 0; i < j; i++ {
		offset += r.Config.Output.Offset
	}
	return offset
}

// GetOutputFilePrefix return output fila path prefix.
func (r *QueryRunner) GetOutputFilePrefix(outputDir, outputFilePrefix string) (string, error) {
	var fp string
	var err error
	if outputDir == "" {
		outputDir, err = ioutil.TempDir("", "esqrunner-")
		if err != nil {
			return fp, err
		}
	}
	if outputFilePrefix == "" {
		chars := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
		length := 8
		var b strings.Builder
		for i := 0; i < length; i++ {
			b.WriteRune(chars[rand.Intn(len(chars))])
		}
		outputFilePrefix = b.String() + "_"
	}
	return filepath.Join(outputDir, outputFilePrefix), nil
}

// WriteToFiles outputs data in various formats, e.g. JSON, CSV, etc.
func (r *QueryRunner) WriteToFiles(fp string) ([]string, error) {
	outputFiles := []string{}
	r.Config.Output.Format = "csv"
	r.Config.Output.Landscape = true
	out, err := r.Output()
	if err != nil {
		return outputFiles, err
	}
	outputFile := fmt.Sprintf("%s_landscape.%s", fp, r.Config.Output.Format)
	if err := writeToFile(outputFile, out); err != nil {
		return outputFiles, err
	}
	outputFiles = append(outputFiles, outputFile)

	r.Config.Output.Landscape = false
	out, err = r.Output()
	if err != nil {
		return outputFiles, err
	}
	outputFile = fmt.Sprintf("%s_portrait.%s", fp, r.Config.Output.Format)
	if err := writeToFile(outputFile, out); err != nil {
		return outputFiles, err
	}
	outputFiles = append(outputFiles, outputFile)

	r.Config.Output.Format = "json"
	out, err = r.Output()
	if err != nil {
		return outputFiles, err
	}
	outputFile = fmt.Sprintf("%s.%s", fp, r.Config.Output.Format)
	if err := writeToFile(outputFile, out); err != nil {
		return outputFiles, err
	}
	outputFiles = append(outputFiles, outputFile)

	r.Config.Output.Format = "js"
	out, err = r.Output()
	if err != nil {
		return outputFiles, err
	}
	outputFile = fmt.Sprintf("%s.%s", fp, r.Config.Output.Format)
	if err := writeToFile(outputFile, out); err != nil {
		return outputFiles, err
	}
	outputFiles = append(outputFiles, outputFile)

	return outputFiles, nil
}
