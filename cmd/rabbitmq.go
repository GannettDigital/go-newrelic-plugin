package cmd

import (
	"github.com/GannettDigital/go-newrelic-plugin/rabbitmq"
	status "github.com/GannettDigital/goStateModule"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(rabbitmqCmd)
}

var rabbitmqCmd = &cobra.Command{
	Use:   "rabbitmq",
	Short: "execute a rabbitmq collection",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("rabbitmq collection")
		rabbitmq.Run(log, prettyPrint, status.GetInfo().Version)
	},
}
