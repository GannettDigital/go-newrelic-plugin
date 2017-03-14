package cmd

import (
	"github.com/GannettDigital/go-newrelic-plugin/nginx"
	status "github.com/GannettDigital/goStateModule"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(nginxCmd)
}

var nginxCmd = &cobra.Command{
	Use:   "nginx",
	Short: "execute an nginx collection",
	Run: func(cmd *cobra.Command, args []string) {
		nginx.Run(log, PrettyPrint, status.GetInfo().Version)
	},
}
