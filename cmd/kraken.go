package cmd

import (
	"github.com/GannettDigital/go-newrelic-plugin/kraken"
	status "github.com/GannettDigital/goStateModule"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(krakenCmd)
}

var krakenCmd = &cobra.Command{
	Use:   "kraken",
	Short: "execute a kraken collection",
	Run: func(cmd *cobra.Command, args []string) {
   log.Info("kraken collection")
		kraken.Run(log, prettyPrint, status.GetInfo().Version)
	},
}
