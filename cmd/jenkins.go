package cmd

import (
	"github.com/GannettDigital/go-newrelic-plugin/jenkins"
	status "github.com/GannettDigital/goStateModule"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(jenkinsCmd)
}

var jenkinsCmd = &cobra.Command{
	Use:   "jenkins",
	Short: "execute a jenkins collection",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("jenkins collection")
		jenkins.Run(log, prettyPrint, status.GetInfo().Version)
	},
}
