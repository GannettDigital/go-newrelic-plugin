package mongo

import (
	"os"

	"github.com/GannettDigital/go-newrelic-plugin/helpers"
	"github.com/Sirupsen/logrus"
)

const NAME string = "mongo"
const EVENT_TYPE string = "DataStoreSample"
const PROVIDER string = "mongo"
const PROTOCOL_VERSION string = "1"

func Run(log *logrus.Logger, prettyPrint bool, version string) {
	// Initialize the output structure
	var data = pluginData{
		Name:            NAME,
		ProtocolVersion: PROTOCOL_VERSION,
		PluginVersion:   version,
		Inventory:       make(map[string]inventoryData),
		Metrics:         make([]metricData, 0),
		Events:          make([]eventData, 0),
	}

	var config = mongoConfig{
		MongoDBUser:     os.Getenv("KEY"),
		MongoDBPassword: os.Getenv("KEY"),
		MongoDBHost:     os.Getenv("KEY"),
		MongoDBPort:     os.Getenv("KEY"),
		MongoDB:         os.Getenv("KEY"),
	}
	validateConfig(log, config)

	var metric = getMetric(log, config)
	data.Metrics = append(data.Metrics, metric)
	fatalIfErr(log, helpers.OutputJSON(data, prettyPrint))
}

func getMetric(log *logrus.Logger, config mongoConfig) map[string]interface{} {
	return map[string]interface{}{
		"event_type": EVENT_TYPE,
		"provider":   PROVIDER,
	}
}

func validateConfig(log *logrus.Logger, config mongoConfig) {
	if config.MongoDBUser == "" || config.MongoDBPassword == "" || config.MongoDBHost == "" || config.MongoDBPort == "" || config.MongoDB == "" {
		log.Fatal("Config Yaml is missing values. Please check the config to continue")
	}
}

func fatalIfErr(log *logrus.Logger, err error) {
	if err != nil {
		log.WithError(err).Fatal("can't continue")
	}
}
