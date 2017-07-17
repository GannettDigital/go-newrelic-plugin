package main

import (
	"fmt"
	"os"

	"github.com/GannettDigital/go-newrelic-plugin/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(-1)
	}

}
