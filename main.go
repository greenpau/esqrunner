package main

import (
	"flag"
	"fmt"
	"github.com/greenpau/versioned"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
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
	client := New()
	flag.StringVar(&configFile, "config", "", "path to configuration file")
	flag.StringVar(&logLevel, "log-level", "info", "logging severity level")
	flag.BoolVar(&isValidate, "validate", false, "validate configuration")
	flag.BoolVar(&isShowVersion, "version", false, "version information")

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

	log.Debugf("configuration file %s", configFile)
	configDir, configFile := filepath.Split(configFile)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AddConfigPath(configDir)
	ext := filepath.Ext(configFile)
	switch ext {
	case "json":
		log.Debugf("configuration syntax is json")
		viper.SetConfigType("json")
	default:
		log.Debugf("configuration syntax is yaml")
		viper.SetConfigType("yml")
	}
	configName := strings.TrimSuffix(configFile, ext)
	log.Debugf("configuration name: %s", configName)
	viper.SetConfigName(configName)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Warnf("error reading configuration file, %s", err)
	}

	if err := viper.Unmarshal(&client.Config); err != nil {
		log.Fatalf("error parsing configuration file: %s", err)
	}

	if isValidate {
		client.ValidateOnly = true
	}

	if err := client.Run(); err != nil {
		log.Fatal(err)
	}
}
