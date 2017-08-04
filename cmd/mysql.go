package cmd

import (
	"github.com/GannettDigital/go-newrelic-plugin/mysql"
	status "github.com/GannettDigital/goStateModule"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(mysqlCmd)
}

var mysqlCmd = &cobra.Command{
	Use:   "mysql",
	Short: "execute a mysql collection",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("mysql collection")
		mysql.Run(log, prettyPrint, status.GetInfo().Version)
	},
}
