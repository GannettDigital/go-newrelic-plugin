package cmd

import (
	"github.com/GannettDigital/go-newrelic-plugin/memcached"
	status "github.com/GannettDigital/goStateModule"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(memcachedCmd)
}

var memcachedCmd = &cobra.Command{
	Use:   "memcached",
	Short: "execute a memcached collection",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("memcached collection")
		memcached.Run(log, prettyPrint, status.GetInfo().Version)
	},
}
