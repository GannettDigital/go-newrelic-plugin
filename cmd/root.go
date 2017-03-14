package cmd

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

var log *logrus.Logger
var PrettyPrint bool

func init() {
	log = logrus.New()
	// Setup logging, redirect logs to stderr and configure the log level.
	log.Out = os.Stderr
	RootCmd.PersistentFlags().BoolVar(&PrettyPrint, "pretty-print", false, "pretty print output")
}

var RootCmd = &cobra.Command{
	Use:   "go-newrelic-plugin",
	Short: "A set of plugins to integrate custom checks into the newrelic infrastructure",
}
