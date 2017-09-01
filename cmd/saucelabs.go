package cmd

import (
	"github.com/GannettDigital/go-newrelic-plugin/saucelabs"
	status "github.com/GannettDigital/goStateModule"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(saucelabsCmd)
}

var saucelabsCmd = &cobra.Command{
	Use:   "saucelabs",
	Short: "execute a saucelabs collection",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("saucelabs collection")
		skel.Run(log, prettyPrint, status.GetInfo().Version)
	},
}
