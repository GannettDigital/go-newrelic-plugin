package mongo

import (
	"fmt"
	"os"

	"github.com/GannettDigital/go-newrelic-plugin/helpers"
	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
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
		MongoDBUser:     os.Getenv("MONGODB_USER"),
		MongoDBPassword: os.Getenv("MONGODB_PASSWORD"),
		MongoDBHost:     os.Getenv("MONGODB_HOST"),
		MongoDBPort:     os.Getenv("MONGODB_PORT"),
		MongoDB:         os.Getenv("MONGODB_DB"),
	}
	validateConfig(log, config)
	mongoURL := fmt.Sprintf("mongodb://%v:%v@%v:%v/%v", config.MongoDBUser, config.MongoDBPassword, config.MongoDBHost, config.MongoDBPort, config.MongoDB)
	session, err := mgo.Dial(mongoURL)
	fatalIfErr(log, err)
	databaseNames, err := session.DatabaseNames()
	fatalIfErr(log, err)
	databaseStatsArray := make([]dbStats, len(databaseNames))
	for index, databaseName := range databaseNames {
		currentDatabase := session.DB(databaseName)
		currentDatabase.Run("dbStats", &databaseStatsArray[index])
	}

	var serverStatusResult serverStatus
	err = session.Run("serverStatus", &serverStatusResult)
	fatalIfErr(log, err)
	data.Metrics = append(data.Metrics, formatServerStatsStructToMap(serverStatusResult))
	for _, databaseStatsStruct := range databaseStatsArray {
		data.Metrics = append(data.Metrics, formatDBStatsStructToMap(databaseStatsStruct))
	}
	fatalIfErr(log, helpers.OutputJSON(data, prettyPrint))
}

func validateConfig(log *logrus.Logger, config mongoConfig) {
	if config.MongoDBUser == "" || config.MongoDBPassword == "" || config.MongoDBHost == "" || config.MongoDBPort == "" || config.MongoDB == "" {
		log.Error(config)
		log.Fatal("Config Yaml is missing values. Please check the config to continue")
	}
}

func fatalIfErr(log *logrus.Logger, err error) {
	if err != nil {
		log.WithError(err).Fatal("can't continue")
	}
}
