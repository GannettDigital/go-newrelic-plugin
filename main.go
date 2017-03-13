package main

import (
	"fmt"
	"os"

	"github.com/GannettDigital/go-newrelic-plugin/couchbase"
	"github.com/GannettDigital/go-newrelic-plugin/nginx"
	"github.com/GannettDigital/go-newrelic-plugin/rabbitmq"
	"github.com/GannettDigital/go-newrelic-plugin/types"
	status "github.com/GannettDigital/goStateModule"
	"github.com/Sirupsen/logrus"
	flags "github.com/jessevdk/go-flags"
)

var log *logrus.Logger
var opts types.Opts
var version string

// list of available collectors.
// these are matched to the --type flag to determine what collector to fire for a given config
var collectors = map[string]types.Collector{
	"nginx":     nginx.Run,
	"couchbase": couchbase.Run,
	"rabbitmq":  rabbitmq.Run,
}

func init() {
	log = logrus.New()
	// Setup logging, redirect logs to stderr and configure the log level.
	log.Out = os.Stderr

	if status.GetInfo().Version == "" {
		version = "0.0.0"
	}
	_, err := flags.Parse(&opts)
	if err != nil {
		if !opts.Version {
			os.Exit(1)
		}
	}

	if opts.Version {
		fmt.Println(fmt.Sprintf("version: %s \nbuilt at: %s", version, status.GetInfo().BuiltAt))
		os.Exit(1)
	}

	if opts.ListTypes {
		ListTypes()
		os.Exit(1)
	}
	if opts.Verbose {
		log.Level = logrus.DebugLevel
	} else {
		log.Level = logrus.InfoLevel
	}

}

func main() {

	typeFound := false
	// main routine
	for name, collector := range collectors {
		if opts.Type == name {
			log.Info("Collector Type: ", name)
			typeFound = true
			collector(log, opts, version)
		}
	}

	if !typeFound {
		log.Error(fmt.Sprintf("collector %s not found!", opts.Type))
		ListTypes()
		os.Exit(1)
	}

}

func ListTypes() {
	concat := ""
	for name, _ := range collectors {
		concat = concat + name + " "
	}
	log.Info("available types: ", concat)

}
