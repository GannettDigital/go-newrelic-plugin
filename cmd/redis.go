package cmd

import (
	"os"

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
		var redisConf = redis.Config{
			RedisHost: os.Getenv("REDISHOST"),
			RedisPort: os.Getenv("REDISPORT"),
			RedisPass: os.Getenv("REDISPASS"),
			RedisDB:   os.Getenv("REDISDB"),
		}
		redis.ValidateConfig(log, &redisConf)
		redis.Run(log, redis.InitRedisClient(redisConf), redisConf, prettyPrint, status.GetInfo().Version)
	},
}
