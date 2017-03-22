package haproxy

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/GannettDigital/paas-api-utils/utilsHTTP"
	"github.com/Sirupsen/logrus"
)

var runner utilsHTTP.HTTPRunner

const NAME string = "haproxy"
const PROVIDER string = "haproxy"
const PROTOCOL_VERSION string = "1"

//HaproxyConfig is the keeper of the config
type HaproxyConfig struct {
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

// OutputJSON takes an object and prints it as a JSON string to the stdout.
// If the pretty attribute is set to true, the JSON will be idented for easy reading.
func OutputJSON(data interface{}, pretty bool) error {
	var output []byte
	var err error

	if pretty {
		output, err = json.MarshalIndent(data, "", "\t")
	} else {
		output, err = json.Marshal(data)
	}

	if err != nil {
		return fmt.Errorf("Error outputting JSON: %s", err)
	}

	if string(output) == "null" {
		fmt.Println("[]")
	} else {
		fmt.Println(string(output))
	}

	return nil
}

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

	var config = HaproxyConfig{
		HaproxyPort:      os.Getenv("HAPROXYPORT"),
		HaproxyStatusURI: os.Getenv("HAPROXYSTATUSURI"),
		HaproxyHost:      os.Getenv("HAPROXYHOST"),
	}
	validateConfig(log, config)

	var metric = getHaproxyStatus(log, config)
	data.Metrics = append(data.Metrics, metric)
	fatalIfErr(log, OutputJSON(data, prettyPrint))
}

func initStats(log *logrus.Logger, config HaproxyConfig) ([][]string, error) {
	haproxyStatsURI := fmt.Sprintf("%v:%v/%v;csv", config.HaproxyHost, config.HaproxyPort, config.HaproxyStatusURI)
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
		return err
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

func getHaproxyStatus(log *logrus.Logger, config HaproxyConfig) ([]MetricData, error) {
	InitialStats, err := initStats(log, config)
	if err != nil {
		log.WithFields(logrus.Fields{
			"haproxyConfig": config,
			"error":         err,
		}).Error("Encountered error querying Stats")
		return make([]MetricData, 0), err
	}
	Stats := make([]MetricData, 0)
	for _, record := range InitialStats {
		if record[0] == "http_frontend" {
			Stats = append(Stats, MetricData{
				"haproxy.frontend.session.current":  toInt(record[4]),
				"haproxy.frontend.session.max":      toInt(record[5]),
				"haproxy.frontend.session.limit":    toInt(record[6]),
				"haproxy.frontend.session.total":    toInt(record[7]),
				"haproxy.frontend.bytes.in_rate":    toInt64(record[8]),
				"haproxy.frontend.bytes.out_rate":   toInt64(record[9]),
				"haproxy.frontend.denied.req_rate":  toInt(record[10]),
				"haproxy.frontend.denied.resp_rate": toInt(record[11]),
				"haproxy.frontend.errors.req_rate":  toInt(record[12]),
				"haproxy.frontend.session.rate":     toInt(record[33]),
				"haproxy.frontend.response.1xx":     toInt(record[39]),
				"haproxy.frontend.response.2xx":     toInt(record[40]),
				"haproxy.frontend.response.3xx":     toInt(record[41]),
				"haproxy.frontend.response.4xx":     toInt(record[42]),
				"haproxy.frontend.response.5xx":     toInt(record[43]),
				"haproxy.frontend.response.other":   toInt(record[44]),
				"haproxy.frontend.requests.rate":    toInt(record[46]),
			})
		} else if record[0] != "stats" && record[1] == "BACKEND" {
			Stats = append(Stats, MetricData{
				"haproxy.backend.queue.current":       toInt(record[2]),
				"haproxy.backend.queue.max":           toInt(record[3]),
				"haproxy.backend.session.current":     toInt(record[4]),
				"haproxy.backend.session.max":         toInt(record[5]),
				"haproxy.backend.session.limit":       toInt(record[6]),
				"haproxy.backend.session.total":       toInt(record[7]),
				"haproxy.backend.bytes.in_rate":       toInt64(record[8]),
				"haproxy.backend.bytes.out_rate":      toInt64(record[9]),
				"haproxy.backend.denied.req_rate":     toInt(record[10]),
				"haproxy.backend.denied.resp_rate":    toInt(record[11]),
				"haproxy.backend.errors.con_rate":     toInt(record[13]),
				"haproxy.backend.errors.resp_rate":    toInt(record[14]),
				"haproxy.backend.warnings.retr_rate":  toInt(record[15]),
				"haproxy.backend.warnings.redis_rate": toInt(record[16]),
				"haproxy.backend.session.rate":        toInt(record[33]),
				"haproxy.backend.response.1xx":        toInt(record[39]),
				"haproxy.backend.response.2xx":        toInt(record[40]),
				"haproxy.backend.response.3xx":        toInt(record[41]),
				"haproxy.backend.response.4xx":        toInt(record[42]),
				"haproxy.backend.response.5xx":        toInt(record[43]),
				"haproxy.backend.response.other":      toInt(record[44]),
				"haproxy.backend.queue.time":          toInt(record[58]),
				"haproxy.backend.connect.time":        toInt(record[59]),
				"haproxy.backend.response.time":       toInt(record[60]),
				"haproxy.backend.session.time":        toInt(record[61]),
			})
		}
	}
	//return Stats, nil
	return Stats, nil
}

func toInt(value string) int {
	if value == "" {
		return 0
	} else {
		valueInt, err := strconv.Atoi(value)
		if err != nil {
			log.WithFields(logrus.Fields{
				"valueInt": valueInt,
				"error":    err,
			}).Error("Error converting value to int")
			return 0
		}
		return valueInt
	}
}

func toInt64(value string) int64 {
	if value == "" {
		return 0
	} else {
		valueInt, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			log.WithFields(logrus.Fields{
				"valueInt": valueInt,
				"error":    err,
			}).Error("Error converting value to int")
			return 0
		}
		return valueInt
	}
}

func validateConfig(log *logrus.Logger, config HaproxyConfig) {
	if config.HaproxyStatusURI == "" {
		log.Fatal("Config Yaml is missing values. Please check the config to continue")
	}
}

func fatalIfErr(log *logrus.Logger, err error) {
	if err != nil {
		log.WithError(err).Fatal("can't continue")
	}
}
