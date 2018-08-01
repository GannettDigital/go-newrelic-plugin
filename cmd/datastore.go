package cmd

import (
	"github.com/GannettDigital/goStateModule"
	"github.com/spf13/cobra"
	"github.com/GannettDigital/go-newrelic-plugin/datastore"
)

func init() {
	RootCmd.AddCommand(datastoreCmd)
}

var datastoreCmd = &cobra.Command{
	Use:   "datastore",
	Short: "execute a datastore real time metric collection",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("datastore collection")
		datastore.Run(log, prettyPrint, status.GetInfo().Version)
	},
}
