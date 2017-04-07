package kraken

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/GannettDigital/paas-api-utils/utilsHTTP"
	"github.com/Sirupsen/logrus"
)

var runner utilsHTTP.HTTPRunner

// NAME - name of plugin
const NAME string = "kraken"

// PROVIDER -
const PROVIDER string = "kraken" //we might want to make this an env tied to kraken version or app name maybe...

// ProtocolVersion -
const ProtocolVersion string = "1"

//KrakenConfig is the keeper of the config
type Config struct {
	KrakenListenPort string
	KrakenStatusURI  string
	KrakenHost       string
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

func init() {
	runner = utilsHTTP.HTTPRunnerImpl{}
}

func Run(log *logrus.Logger, prettyPrint bool, version string) {

	// Initialize the output structure
	var data = PluginData{
		Name:            NAME,
		ProtocolVersion: ProtocolVersion,
		PluginVersion:   version,
		Inventory:       make(map[string]InventoryData),
		Metrics:         make([]MetricData, 0),
		Events:          make([]EventData, 0),
	}

	var krakenConf = Config{
		KrakenListenPort: os.Getenv("KRAKEN_PORT"),
		KrakenHost:       os.Getenv("KRAKEN_HOST"),
		KrakenStatusURI:  os.Getenv("KRAKEN_STATUSURI"),
	}
	validateConfig(log, krakenConf)

	var metric = scrapeStatus(log, getKrakenStatus(log, krakenConf))

	data.Metrics = append(data.Metrics, metric)
	fatalIfErr(log, OutputJSON(data, prettyPrint))
}

func validateConfig(log *logrus.Logger, krakenConf Config) {
	if krakenConf.KrakenHost == "" || krakenConf.KrakenListenPort == "" || krakenConf.KrakenStatusURI == "" {
		log.Fatal("Config Yaml is missing values. Please check the config to continue")
	}
}

func fatalIfErr(log *logrus.Logger, err error) {
	if err != nil {
		log.WithError(err).Fatal("can't continue")
	}
}

func getKrakenStatus(log *logrus.Logger, config Config) string {
	krakenStatus := fmt.Sprintf("%v:%v/%v", config.KrakenHost, config.KrakenListenPort, config.KrakenStatusURI)
	httpReq, err := http.NewRequest("GET", krakenStatus, bytes.NewBuffer([]byte("")))
	// http.NewRequest error
	fatalIfErr(log, err)
	code, data, err := runner.CallAPI(log, nil, httpReq, &http.Client{})
	if err != nil || code != 200 {
		log.WithFields(logrus.Fields{
			"code":                   code,
			"data":                   string(data),
			"httpReq":                httpReq,
			"config.KrakenStatusPage": config.KrakenHost,
			"config.KrakenListenPort": config.KrakenListenPort,
			"config.KrakenStatusURI":  config.KrakenStatusURI,
			"error":                  err,
		}).Fatal("Encountered error calling CallAPI")
		return ""
	}

	return string(data)
}

func scrapeStatus(log *logrus.Logger, status string) map[string]interface{} {

	multi = regexp.MustCompile(`Version: (\s)`).FindString(status)
	contents = strings.Fields(multi)
	krakenVersion := contents[0]

	multi = regexp.MustCompile(`Customer: (\s)`).FindString(status)
	contents = strings.Fields(multi)
	krakenCustomer := contents[0]

	multi = regexp.MustCompile(`Project: (\s)`).FindString(status)
	contents = strings.Fields(multi)
	krakenProject := contents[0]

	multi = regexp.MustCompile(`State: (\s)`).FindString(status)
	contents = strings.Fields(multi)
	krakenState := contents[0]

	multi := regexp.MustCompile(`Samples count: (\d+), (\d+(\.\d+)?). failures`).FindString(status)
	contents := strings.Fields(multi)
	sample_count := contents[0]
	sample_failure := contents[1]

	multi = regexp.MustCompile(`Average times: total (\d+(\.\d+)?), latency (\d+(\.\d+)?), connect (\d+(\.\d+)?)`).FindString(status)
	contents = strings.Fields(multi)
	avg_resp_time := contents[0]
	avg_latency := contents[1]
	avg_conn_time := contents[2]

	multi = regexp.MustCompile(`Percentile  50.0.: (\d+(\.\d+)?)`).FindString(status)
	contents = strings.Fields(multi)
	percentiles_50 := contents[0]

	multi = regexp.MustCompile(`Percentile  90.0.: (\d+(\.\d+)?)`).FindString(status)
	contents = strings.Fields(multi)
	percentiles_90 := contents[0]

	multi = regexp.MustCompile(`Percentile  95.0.: (\d+(\.\d+)?)`).FindString(status)
	contents = strings.Fields(multi)
	percentiles_95 := contents[0]

	multi = regexp.MustCompile(`Percentile  99.0.: (\d+(\.\d+)?)`).FindString(status)
	contents = strings.Fields(multi)
	percentiles_99 := contents[0]

	multi = regexp.MustCompile(`Percentile  100.0.: (\d+(\.\d+)?)`).FindString(status)
	contents = strings.Fields(multi)
	percentiles_100 := contents[0]

	log.WithFields(logrus.Fields{
		"kraken_version":  krakenVersion,
		"kraken_customer":  krakenCustomer,
		"kraken_project":  krakenProject,
		"kraken_state":  krakenState,
		"avg_resp_time":  avg_resp_time,
		"avg_latency":  avg_latency,
		"avg_conn_time":  avg_conn_time,
		"sample_count": sample_count,
		"sample_failure": sample_failure,
		"avg_rt":  avg_rt,
		"duration":  duration,
	}).Debugf("Scraped KRAKEN values")
	return map[string]interface{}{
		"event_type":               "KrakenSample",
		"provider":                 PROVIDER,
		"kraken.version":						krakenVersion
		"kraken.customer":			   	krakenCustomer
		"kraken.project":						krakenProject
		"kraken.state":				   		krakenState
		"kraken.kpi.avg_resp_time": toInt(log, avg_resp_time),
		"kraken.kpi.avg_latency":   toInt(log, avg_latency),
		"kraken.kpi.avg_conn_time": toInt(log, avg_conn_time),
		"kraken.kpi.percentiles.50":   toInt(log, percentiles_50),
		"kraken.kpi.percentiles.90":   toInt(log, percentiles_90),
		"kraken.kpi.percentiles.95":   toInt(log, percentiles_95),
		"kraken.kpi.percentiles.99":   toInt(log, percentiles_99),
		"kraken.kpi.percentiles.100":   toInt(log, percentiles_100),
		"kraken.sample_count":      toInt(log, sample_count),
		"kraken.sample_failure":    toInt(log, sample_failure),
		"kraken.avg_rt":            toInt(log, avg_rt),
		"kraken.duration":          toInt(log, duration),
	}
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
