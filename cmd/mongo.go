package cmd

import (
	"os"

	"github.com/GannettDigital/go-newrelic-plugin/mongo"
	status "github.com/GannettDigital/goStateModule"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(mongoCmd)
}

var mongoCmd = &cobra.Command{
	Use:   "mongo",
	Short: "execute a mongo collection",
	Run: func(cmd *cobra.Command, args []string) {
		var config = mongo.Config{
			MongoDBUser:     os.Getenv("MONGODB_USER"),
			MongoDBPassword: os.Getenv("MONGODB_PASSWORD"),
			MongoDBHost:     os.Getenv("MONGODB_HOST"),
			MongoDBPort:     os.Getenv("MONGODB_PORT"),
			MongoDB:         os.Getenv("MONGODB_DB"),
		}
		mongo.ValidateConfig(log, config)
		mongo.Run(log, mongo.InitMongoClient(log, config), config, prettyPrint, status.GetInfo().Version)
	},
}
