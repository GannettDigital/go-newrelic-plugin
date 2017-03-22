package mongo

import (
	"fmt"

	"github.com/GannettDigital/go-newrelic-plugin/helpers"
	"github.com/Sirupsen/logrus"
)

const NAME string = "mongo"
const EVENT_TYPE string = "DataStoreSample"
const PROVIDER string = "mongo"
const PROTOCOL_VERSION string = "1"

func Run(log *logrus.Logger, session Session, mongoConfig Config, prettyPrint bool, version string) {
	// Initialize the output structure
	var data = pluginData{
		Name:            NAME,
		ProtocolVersion: PROTOCOL_VERSION,
		PluginVersion:   version,
		Inventory:       make(map[string]inventoryData),
		Metrics:         make([]metricData, 0),
		Events:          make([]eventData, 0),
	}

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

// InitMongoClient - function to create a mongo client
func InitMongoClient(log *logrus.Logger, config Config) Session {
	mongoURL := fmt.Sprintf("mongodb://%v:%v@%v:%v/%v", config.MongoDBUser, config.MongoDBPassword, config.MongoDBHost, config.MongoDBPort, config.MongoDB)
	return NewSession(mongoURL)
}

func ValidateConfig(log *logrus.Logger, config Config) {
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
