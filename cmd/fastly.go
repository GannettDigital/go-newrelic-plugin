package cmd

import (
	"github.com/GannettDigital/go-newrelic-plugin/fastly"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(fastlyCmd)
}

var fastlyCmd = &cobra.Command{
	Use:   "fastly",
	Short: "execute a fastly real time metric collection",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("fastly collection")
		fastly.Run(log, prettyPrint, "1")
	},
}
