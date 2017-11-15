package cmd

import (
	"github.com/GannettDigital/go-newrelic-plugin/jira"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(jiraCmd)
}

var jiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "execute a jira collector",
	Run: func(cmd *cobra.Command, args []string) {
		jira.Run(log)
	},
}
