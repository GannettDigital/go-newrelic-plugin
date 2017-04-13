package cmd

import (
	"os"

	"github.com/GannettDigital/go-newrelic-plugin/helpers"
	"github.com/GannettDigital/go-newrelic-plugin/sslCheck"
	status "github.com/GannettDigital/goStateModule"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(sslCheckCmd)
}

var sslCheckCmd = &cobra.Command{
	Use:   "sslCheck",
	Short: "Records events based on host certificate expirations",
	Run: func(cmd *cobra.Command, args []string) {
		expiredEventPeriod, err := helpers.ToInt(os.Getenv("SSLCHECK_EXPIRED_EVENT_PERIOD"))
		if err != nil {
			log.Fatalf("Invaliding Integer for event period: %v\n", err)
		}
		var config = sslCheck.Config{
			Hosts:              os.Getenv("SSLCHECK_HOSTS"),
			ExpiredEventPeriod: expiredEventPeriod,
		}
		err = sslCheck.ValidateConfig(config)
		if err != nil {
			log.Fatalf("invalid config: %v\n", err)
		}
		sslCheck.Run(log, config, prettyPrint, status.GetInfo().Version)
	},
}
