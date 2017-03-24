package cmd

import (
	"github.com/GannettDigital/go-newrelic-plugin/haproxy"
	status "github.com/GannettDigital/goStateModule"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(haproxyCmd)
}

var haproxyCmd = &cobra.Command{
	Use:   "haproxy",
	Short: "execute a haproxy collection",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("haproxy collection")
		haproxy.Run(log, prettyPrint, status.GetInfo().Version)
	},
}
