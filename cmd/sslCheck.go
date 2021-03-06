package cmd

import (
	"io/ioutil"
	"os"

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
		rootCaFile := os.Getenv("SSLCHECK_ROOT_CAS")
		var rootCAPem []byte
		var err error
		if rootCaFile != "" {
			rootCAPem, err = ioutil.ReadFile(rootCaFile)
			if err != nil {
				log.Fatalf("Error Reading Ca File: %v\n", err)
			}
		}

		hosts, err := sslCheck.ProcessHosts(os.Getenv("SSLCHECK_HOSTS"))
		if err != nil {
			log.Fatalf("Error Processing Hosts: %v\n", err)
		}
		var config = sslCheck.Config{
			Hosts: hosts,
		}
		err = sslCheck.ValidateConfig(config)
		if err != nil {
			log.Fatalf("invalid config: %v\n", err)
		}
		sslCheck.Run(log, config, rootCAPem, prettyPrint, status.GetInfo().Version)
	},
}
