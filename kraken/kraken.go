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
	}
	validateConfig(log, krakenConf)

	var metric = scrapeStatus(log, getKrakenStatus(log, krakenConf))

	data.Metrics = append(data.Metrics, metric)
	fatalIfErr(log, OutputJSON(data, prettyPrint))
}

func validateConfig(log *logrus.Logger, krakenConf Config) {
	if krakenConf.KrakenHost == "" || krakenConf.KrakenListenPort == ""{
		log.Fatal("Config Yaml is missing values. Please check the config to continue")
	}
}

func fatalIfErr(log *logrus.Logger, err error) {
	if err != nil {
		log.WithError(err).Fatal("can't continue")
	}
}

func getKrakenStatus(log *logrus.Logger, config Config) string {
	krakenStatus := fmt.Sprintf("%v:%v/", config.KrakenHost, config.KrakenListenPort)
	httpReq, err := http.NewRequest("GET", krakenStatus, bytes.NewBuffer([]byte("")))
	// http.NewRequest error
	fatalIfErr(log, err)
	code, data, err := runner.CallAPI(log, nil, httpReq, &http.Client{})
	if err != nil || code != 200 {
		log.WithFields(logrus.Fields{
			"code":                    code,
			"data":                    string(data),
			"httpReq":                 httpReq,
			"config.KrakenStatusPage": config.KrakenHost,
			"config.KrakenListenPort": config.KrakenListenPort,
			"error":                   err,
		}).Fatal("Encountered error calling CallAPI")
		return ""
	}

	return string(data)
}

func scrapeStatus(log *logrus.Logger, status string) map[string]interface{} {
	var replacer = strings.NewReplacer(",", "", "\t", "")

	multi := regexp.MustCompile(`Version: (\d+(\.\d+)?(\.\d+)?)`).FindString(status)
	contents := strings.Fields(multi)
	krakenVersion := contents[1]

	multi = regexp.MustCompile(`Customer: (\w+)`).FindString(status)
	contents = strings.Fields(multi)
	krakenCustomer := contents[1]

	multi = regexp.MustCompile(`Project: (\w+)`).FindString(status)
	contents = strings.Fields(multi)
	krakenProject := contents[1]

	multi = regexp.MustCompile(`State: (\w+)`).FindString(status)
	contents = strings.Fields(multi)
	krakenState := contents[1]

	multi = regexp.MustCompile(`Samples count: (\d+), (\d+(\.\d+)?). failures`).FindString(status)
	contents = strings.Fields(multi)
	sample_count := replacer.Replace(contents[2])
	sample_failure := contents[3]

	multi = regexp.MustCompile(`Test duration: (.*)`).FindString(status)
	contents = strings.Fields(multi)
	duration := contents[2]

	multi = regexp.MustCompile(`Average times: total (\d+(\.\d+)?), latency (\d+(\.\d+)?), connect (\d+(\.\d+)?)`).FindString(status)
	contents = strings.Fields(multi)
	avg_resp_time := replacer.Replace(contents[3])
	avg_latency := replacer.Replace(contents[5])
	avg_conn_time := contents[7]

	multi = regexp.MustCompile(`Percentile\s+50.0%:\s+(\d+(\.\d+)?).`).FindString(status)
	contents = strings.Fields(multi)
	percentiles_50 := contents[2]

	multi = regexp.MustCompile(`Percentile\s+90.0.:\s+(\d+(\.\d+)?).`).FindString(status)
	contents = strings.Fields(multi)
	percentiles_90 := contents[2]

	multi = regexp.MustCompile(`Percentile\s+95.0.:\s+(\d+(\.\d+)?).`).FindString(status)
	contents = strings.Fields(multi)
	percentiles_95 := contents[2]

	multi = regexp.MustCompile(`Percentile\s+99.0.:\s+(\d+(\.\d+)?).`).FindString(status)
	contents = strings.Fields(multi)
	percentiles_99 := contents[2]

	multi = regexp.MustCompile(`Percentile\s+100.0.:\s+(\d+(\.\d+)?).`).FindString(status)
	contents = strings.Fields(multi)
	percentiles_100 := contents[2]

	log.WithFields(logrus.Fields{
		"kraken_version":  krakenVersion,
		"kraken_customer":  krakenCustomer,
		"kraken_project":  krakenProject,
		"kraken_state":  krakenState,
		"avg_resp_time":  avg_resp_time,
		"avg_latency":  avg_latency,
		"avg_conn_time":  avg_conn_time,
		"percentiles_50":  percentiles_50,
		"percentiles_90":  percentiles_90,
		"percentiles_95":  percentiles_95,
		"percentiles_99":  percentiles_99,
		"percentiles_100":  percentiles_100,
		"sample_count": sample_count,
		"sample_failure": sample_failure,
		"duration":  duration,
	}).Debugf("Scraped KRAKEN values")
	return map[string]interface{}{
		"event_type":               	 "GKrakenSample",
		"provider":                 	 PROVIDER,
		"kraken.version":							 krakenVersion,
		"kraken.customer":			   	   krakenCustomer,
		"kraken.project":							 krakenProject,
		"kraken.state":				   			 krakenState,
		"kraken.kpi.avg_resp_time": 	 avg_resp_time,
		"kraken.kpi.avg_latency":   	 avg_latency,
		"kraken.kpi.avg_conn_time": 	 avg_conn_time,
		"kraken.kpi.percentiles.50":   percentiles_50,
		"kraken.kpi.percentiles.90":   percentiles_90,
		"kraken.kpi.percentiles.95":   percentiles_95,
		"kraken.kpi.percentiles.99":   percentiles_99,
		"kraken.kpi.percentiles.100":  percentiles_100,
		"kraken.sample_count":      	 sample_count,
		"kraken.sample_failure":   		 sample_failure,
		"kraken.duration":          	 duration,
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
