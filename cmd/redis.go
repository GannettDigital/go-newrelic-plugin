package cmd

import (
	"github.com/GannettDigital/go-newrelic-plugin/redis"
	status "github.com/GannettDigital/goStateModule"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(redisCmd)
}

var redisCmd = &cobra.Command{
	Use:   "redis",
	Short: "execute a redis collection",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("redis collection")
		redis.Run(log, prettyPrint, status.GetInfo().Version)
	},
}
