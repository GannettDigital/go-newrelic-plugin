package cmd

import (
	"github.com/GannettDigital/go-newrelic-plugin/couchbase"
	status "github.com/GannettDigital/goStateModule"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(couchbaseCmd)
}

var couchbaseCmd = &cobra.Command{
	Use:   "couchbase",
	Short: "execute a couchbase collection",
	Run: func(cmd *cobra.Command, args []string) {
		couchbase.Run(log, PrettyPrint, status.GetInfo().Version)
	},
}
