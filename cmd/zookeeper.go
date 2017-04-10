package cmd

import (
	"github.com/GannettDigital/go-newrelic-plugin/zookeeper"
	status "github.com/GannettDigital/goStateModule"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(zookeeperCmd)
}

var zookeeperCmd = &cobra.Command{
	Use:   "zookeeper",
	Short: "execute a zookeeper collection",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("zookeeper collection")
		zookeeper.Run(log, prettyPrint, status.GetInfo().Version)
	},
}
