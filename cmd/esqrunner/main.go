package main

import (
	"flag"
	"fmt"
	"github.com/greenpau/esqrunner"
	"github.com/greenpau/versioned"
	log "github.com/sirupsen/logrus"
	"os"
)

var (
	app        *versioned.PackageManager
	appVersion string
	gitBranch  string
	gitCommit  string
	buildUser  string
	buildDate  string
)

func init() {
	app = versioned.NewPackageManager("esqrunner")
	app.Description = "Run Elasticsearh queries and create metrics based on the results."
	app.Documentation = "https://github.com/greenpau/esqrunner/"
	app.SetVersion(appVersion, "")
	app.SetGitBranch(gitBranch, "")
	app.SetGitCommit(gitCommit, "")
	app.SetBuildUser(buildUser, "")
	app.SetBuildDate(buildDate, "")
}

func main() {
	var configFile string
	var logLevel string
	var isShowVersion bool
	var isValidate bool
	var datePicker string
	var isLandscape bool
	client := esqrunner.New()
	flag.StringVar(&configFile, "config", "", "path to configuration file")
	flag.StringVar(&logLevel, "log-level", "info", "logging severity level")
	flag.BoolVar(&isValidate, "validate", false, "validate configuration")
	flag.BoolVar(&isShowVersion, "version", false, "version information")
	flag.StringVar(&datePicker, "datepicker", "", "date pattern, e.g. last 7 days, interval 1 day")
	flag.BoolVar(&isLandscape, "landscape", false, "landscape output")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\n%s - %s\n\n", app.Name, app.Description)
		fmt.Fprintf(os.Stderr, "Usage: %s [arguments]\n\n", app.Name)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nDocumentation: %s\n\n", app.Documentation)
	}
	flag.Parse()
	if isShowVersion {
		fmt.Fprintf(os.Stdout, "%s\n", app.Banner())
		os.Exit(0)
	}

	if logLevel != "" {
		if level, err := log.ParseLevel(logLevel); err == nil {
			log.SetLevel(level)
		} else {
			log.Fatalf("%s", err.Error())
		}
	}

	//log.SetFormatter(&log.JSONFormatter{})

	if configFile == "" {
		log.Fatalf("no configuration file")
	}

	if err := client.ReadInConfig(configFile); err != nil {
		log.Warnf("error reading configuration file, %s", err)
	}

	if isValidate {
		client.ValidateOnly = true
		if err := client.ValidateConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "invalid config: %s\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "configuration is valid\n")
		os.Exit(0)
	}

	if datePicker == "" {
		log.Fatalf("datepicker is required")
	}

	if err := client.Config.AddDates(datePicker); err != nil {
		log.Fatalf("invalid dates: %s", err)
	}

	client.Config.Output.Landscape = isLandscape

	if err := client.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
