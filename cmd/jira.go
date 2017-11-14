package cmd

import (
	"github.com/GannettDigital/go-newrelic-plugin/jira"
	status "github.com/GannettDigital/goStateModule"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(jiraCmd)
}

var jiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "execute a jira collector",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("collecting jira metrics")
		jira.Run(log, prettyPrint, status.GetInfo().Version)
	},
}
