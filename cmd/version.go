package cmd

import (
	"fmt"

	status "github.com/GannettDigital/goStateModule"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

var version string
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of go-newrelic-plugin",
	Run: func(cmd *cobra.Command, args []string) {
		if status.GetInfo().Version == "" {
			version = "0.0.0"
		}
		fmt.Println(fmt.Sprintf("version: %s \nbuilt at: %s", version, status.GetInfo().BuiltAt))
	},
}
