package cmd

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

var log *logrus.Logger
var prettyPrint bool
var verbose bool

func init() {
	log = logrus.New()
	// Setup logging, redirect logs to stderr and configure the log level.
	log.Out = os.Stderr
	RootCmd.PersistentFlags().BoolVar(&prettyPrint, "pretty-print", false, "pretty print output")
	RootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "verbose output")

	if verbose {
		log.Level = logrus.DebugLevel
	}
}

var RootCmd = &cobra.Command{
	Use:   "go-newrelic-plugin",
	Short: "A set of plugins to integrate custom checks into the newrelic infrastructure",
}
