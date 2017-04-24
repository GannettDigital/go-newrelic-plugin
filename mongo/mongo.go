package mongo

import (
	"errors"
	"fmt"

	"github.com/GannettDigital/go-newrelic-plugin/helpers"
	"github.com/Sirupsen/logrus"
)

const NAME string = "mongo"
const EVENT_TYPE string = "DatastoreSample"
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

	databaseStatsArray := readDBStats(log, session)
	for _, databaseStatsStruct := range databaseStatsArray {
		data.Metrics = append(data.Metrics, formatDBStatsStructToMap(databaseStatsStruct))
	}

	replEnabled, databaseReplicatStats := readDBReplicaStats(log, session.DB("admin"))
	if replEnabled {
		for index := range databaseReplicatStats.Members {
			data.Metrics = append(data.Metrics, formatReplStatsStructToMap(databaseReplicatStats, index))
		}
	}

	serverStatusResult := readServerStats(log, session)
	data.Metrics = append(data.Metrics, formatServerStatsStructToMap(serverStatusResult))

	fatalIfErr(log, helpers.OutputJSON(data, prettyPrint))
}

func readServerStats(log *logrus.Logger, session Session) serverStatus {
	var serverStatusResult serverStatus
	err := session.Run("serverStatus", &serverStatusResult)
	fatalIfErr(log, err)
	return serverStatusResult
}

func readDBStats(log *logrus.Logger, session Session) []dbStats {
	databaseNames, err := session.DatabaseNames()
	fatalIfErr(log, err)
	databaseStatsArray := make([]dbStats, len(databaseNames))
	for index, databaseName := range databaseNames {
		currentDatabase := session.DB(databaseName)
		dbStatsErr := currentDatabase.Run("dbStats", &databaseStatsArray[index])
		if dbStatsErr != nil {
			err = dbStatsErr
		}
	}
	fatalIfErr(log, err)
	return databaseStatsArray
}

func readDBReplicaStats(log *logrus.Logger, db DataLayer) (bool, ReplStats) {
	databaseReplicaStats := ReplStats{}
	err := db.Run("replSetGetStatus", &databaseReplicaStats)
	if err != nil {
		if err.Error() == "not running with --replSet" {
			return false, ReplStats{}
		}
		fatalIfErr(log, err)
	}
	return true, databaseReplicaStats
}

// InitMongoClient - function to create a mongo client
func InitMongoClient(log *logrus.Logger, config Config) Session {
	mongoURL := fmt.Sprintf("mongodb://%v:%v@%v:%v/%v", config.MongoDBUser, config.MongoDBPassword, config.MongoDBHost, config.MongoDBPort, config.MongoDB)
	return NewSession(mongoURL)
}

// ValidateConfig validates the config
func ValidateConfig(config Config) error {
	if config.MongoDBUser == "" {
		return errors.New("mongo DBUser must be set")
	}
	if config.MongoDBPassword == "" {
		return errors.New("mongo DBPassword must be set")
	}
	if config.MongoDBHost == "" {
		return errors.New("mongo DBHost must be set")
	}
	if config.MongoDBPort == "" {
		return errors.New("mongo DBPort must be set")
	}
	if config.MongoDB == "" {
		return errors.New("mongo DB must be set")
	}
	return nil
}

func fatalIfErr(log *logrus.Logger, err error) {
	if err != nil {
		log.WithError(err).Fatal("can't continue")
	}
}
