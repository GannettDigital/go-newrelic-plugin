package haproxy

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/GannettDigital/go-newrelic-plugin/helpers"
	"github.com/GannettDigital/paas-api-utils/utilsHTTP"
	"github.com/Sirupsen/logrus"
)

var runner utilsHTTP.HTTPRunner

const NAME string = "haproxy"
const PROVIDER string = "haproxy"
const PROTOCOL_VERSION string = "1"
const EVENT_TYPE string = "LoadBalancerSample"

//Config is the keeper of the config
type Config struct {
	HaproxyPort      string
	HaproxyStatusURI string
	HaproxyHost      string
}

// InventoryData is the data type for inventory data produced by a plugin data
// source and emitted to the agent's inventory data store
type InventoryData map[string]interface{}

// MetricData is the data type for events produced by a plugin data source and
// emitted to the agent's metrics data store
type MetricData map[string]interface{}

// EventData is the data type for single shot events
type EventData map[string]interface{}

// PluginData defines the format of the output JSON that plugins will return
type PluginData struct {
	Name            string                   `json:"name"`
	ProtocolVersion string                   `json:"protocol_version"`
	PluginVersion   string                   `json:"plugin_version"`
	Metrics         []MetricData             `json:"metrics"`
	Inventory       map[string]InventoryData `json:"inventory"`
	Events          []EventData              `json:"events"`
	Status          string                   `json:"status"`
}

func init() {
	runner = &utilsHTTP.HTTPRunnerImpl{}
}

// Run is the entry point for the collector
func Run(log *logrus.Logger, prettyPrint bool, version string) {

	// Initialize the output structure
	var data = PluginData{
		Name:            NAME,
		ProtocolVersion: PROTOCOL_VERSION,
		PluginVersion:   version,
		Inventory:       make(map[string]InventoryData),
		Metrics:         make([]MetricData, 0),
		Events:          make([]EventData, 0),
	}

	var haproxyConf = Config{
		HaproxyPort:      os.Getenv("HAPROXYPORT"),
		HaproxyStatusURI: os.Getenv("HAPROXYSTATUSURI"),
		HaproxyHost:      os.Getenv("HAPROXYHOST"),
	}
	validErr := validateConfig(log, haproxyConf)
	if validErr != nil {
		log.Fatalf("config: %v\n", validErr)
	}

	metric, err := getHaproxyStatus(log, haproxyConf)
	fatalIfErr(log, err)

	data.Metrics = metric
	fatalIfErr(log, helpers.OutputJSON(data, prettyPrint))
}

func initStats(log *logrus.Logger, haproxyConf Config) ([][]string, error) {
	haproxyStatsURI := fmt.Sprintf("%v:%v/%v;csv", haproxyConf.HaproxyHost, haproxyConf.HaproxyPort, haproxyConf.HaproxyStatusURI)
	httpReq, err := http.NewRequest("GET", haproxyStatsURI, bytes.NewBuffer([]byte("")))
	if err != nil {
		log.WithFields(logrus.Fields{
			"haproxyStatsURI": haproxyStatsURI,
			"error":           err,
		}).Error("Encountered error creating http.NewRequest")
		return [][]string{}, err
	}
	code, data, err := runner.CallAPI(log, nil, httpReq, &http.Client{})
	if err != nil || code != 200 {
		log.WithFields(logrus.Fields{
			"code":    code,
			"data":    string(data),
			"httpReq": httpReq,
			"error":   err,
		}).Error("Encountered error calling CallAPI")
		return nil, err
	}
	r := csv.NewReader(strings.NewReader(string(data)))
	everything, err := r.ReadAll()
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("Unable to decode initial CSV stats")
		return everything, err
	}
	return everything, nil
}

func getHaproxyStatus(log *logrus.Logger, haproxyConf Config) ([]MetricData, error) {
	InitialStats, err := initStats(log, haproxyConf)
	if err != nil {
		log.WithFields(logrus.Fields{
			"haproxyConfig": haproxyConf,
			"error":         err,
		}).Error("Encountered error querying Stats")
		return nil, err
	}
	Stats := make([]MetricData, 0)
	for _, record := range InitialStats {
		if record[0] != "stats" && record[1] == "FRONTEND" {
			Stats = append(Stats, MetricData{
				"event_type":                        EVENT_TYPE,
				"provider":                          PROVIDER,
				"haproxy.type":                      "frontend",
				"haproxy.frontend.session.current":  toInt(log, record[4]),
				"haproxy.frontend.session.max":      toInt(log, record[5]),
				"haproxy.frontend.session.limit":    toInt(log, record[6]),
				"haproxy.frontend.session.total":    toInt(log, record[7]),
				"haproxy.frontend.bytes.in_rate":    toInt64(log, record[8]),
				"haproxy.frontend.bytes.out_rate":   toInt64(log, record[9]),
				"haproxy.frontend.denied.req_rate":  toInt(log, record[10]),
				"haproxy.frontend.denied.resp_rate": toInt(log, record[11]),
				"haproxy.frontend.errors.req_rate":  toInt(log, record[12]),
				"haproxy.frontend.session.rate":     toInt(log, record[33]),
				"haproxy.frontend.response.1xx":     toInt(log, record[39]),
				"haproxy.frontend.response.2xx":     toInt(log, record[40]),
				"haproxy.frontend.response.3xx":     toInt(log, record[41]),
				"haproxy.frontend.response.4xx":     toInt(log, record[42]),
				"haproxy.frontend.response.5xx":     toInt(log, record[43]),
				"haproxy.frontend.response.other":   toInt(log, record[44]),
				"haproxy.frontend.requests.rate":    toInt(log, record[46]),
			})
		} else if record[0] != "stats" && record[1] == "BACKEND" {
			Stats = append(Stats, MetricData{
				"event_type":                          EVENT_TYPE,
				"provider":                            PROVIDER,
				"haproxy.type":                        "backend",
				"haproxy.backend.queue.current":       toInt(log, record[2]),
				"haproxy.backend.queue.max":           toInt(log, record[3]),
				"haproxy.backend.session.current":     toInt(log, record[4]),
				"haproxy.backend.session.max":         toInt(log, record[5]),
				"haproxy.backend.session.limit":       toInt(log, record[6]),
				"haproxy.backend.session.total":       toInt(log, record[7]),
				"haproxy.backend.bytes.in_rate":       toInt64(log, record[8]),
				"haproxy.backend.bytes.out_rate":      toInt64(log, record[9]),
				"haproxy.backend.denied.req_rate":     toInt(log, record[10]),
				"haproxy.backend.denied.resp_rate":    toInt(log, record[11]),
				"haproxy.backend.errors.con_rate":     toInt(log, record[13]),
				"haproxy.backend.errors.resp_rate":    toInt(log, record[14]),
				"haproxy.backend.warnings.retr_rate":  toInt(log, record[15]),
				"haproxy.backend.warnings.redis_rate": toInt(log, record[16]),
				"haproxy.backend.session.rate":        toInt(log, record[33]),
				"haproxy.backend.response.1xx":        toInt(log, record[39]),
				"haproxy.backend.response.2xx":        toInt(log, record[40]),
				"haproxy.backend.response.3xx":        toInt(log, record[41]),
				"haproxy.backend.response.4xx":        toInt(log, record[42]),
				"haproxy.backend.response.5xx":        toInt(log, record[43]),
				"haproxy.backend.response.other":      toInt(log, record[44]),
				"haproxy.backend.queue.time":          toInt(log, record[58]),
				"haproxy.backend.connect.time":        toInt(log, record[59]),
				"haproxy.backend.response.time":       toInt(log, record[60]),
				"haproxy.backend.session.time":        toInt(log, record[61]),
			})
		}
	}
	//return Stats, nil
	return Stats, err
}

func toInt(log *logrus.Logger, value string) int {
	if value == "" {
		return 0
	}
	valueInt, err := strconv.Atoi(value)
	if err != nil {
		log.WithFields(logrus.Fields{
			"valueInt": valueInt,
			"error":    err,
		}).Debug("Error converting value to int")

		return 0
	}

	return valueInt
}

func toInt64(log *logrus.Logger, value string) int64 {
	if value == "" {
		return 0
	}
	valueInt, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		log.WithFields(logrus.Fields{
			"valueInt": valueInt,
			"error":    err,
		}).Debug("Error converting value to int")

		return 0
	}

	return valueInt
}

func validateConfig(log *logrus.Logger, haproxyConf Config) error {
	if haproxyConf.HaproxyStatusURI == "" {
		return errors.New("Config is missing the HaproxyStatusURI. Please check the config to continue")
	}
	if haproxyConf.HaproxyPort == "" {
		return errors.New("Config is missing the HaproxyPort for the HTTP status page. Please check the config to continue")
	}
	if haproxyConf.HaproxyHost == "" {
		return errors.New("Config is missing the HaproxyHost. Please check the config to continue")
	}
	return nil
}

func fatalIfErr(log *logrus.Logger, err error) {
	if err != nil {
		log.WithError(err).Fatal("can't continue")
	}
}
