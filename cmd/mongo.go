package cmd

import (
	"github.com/GannettDigital/go-newrelic-plugin/mongo"
	status "github.com/GannettDigital/goStateModule"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(mongoCmd)
}

var mongoCmd = &cobra.Command{
	Use:   "mongo",
	Short: "execute a mongo collection",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("mongo collection")
		mongo.Run(log, prettyPrint, status.GetInfo().Version)
	},
}
