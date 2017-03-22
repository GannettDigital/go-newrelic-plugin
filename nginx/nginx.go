package nginx

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
const NAME string = "nginx"

// PROVIDER -
const PROVIDER string = "nginx" //we might want to make this an env tied to nginx version or app name maybe...

// ProtocolVersion -
const ProtocolVersion string = "1"

//NginxConfig is the keeper of the config
type Config struct {
	NginxListenPort string
	NginxStatusURI  string
	NginxHost       string
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

	var nginxConf = Config{
		NginxListenPort: os.Getenv("NGINXLISTENPORT"),
		NginxHost:       os.Getenv("NGINXHOST"),
		NginxStatusURI:  os.Getenv("NGINXSTATUSURI"),
	}
	validateConfig(log, nginxConf)

	var metric = scrapeStatus(log, getNginxStatus(log, nginxConf))

	data.Metrics = append(data.Metrics, metric)
	fatalIfErr(log, OutputJSON(data, prettyPrint))
}

func validateConfig(log *logrus.Logger, nginxConf Config) {
	if nginxConf.NginxHost == "" || nginxConf.NginxListenPort == "" || nginxConf.NginxStatusURI == "" {
		log.Fatal("Config Yaml is missing values. Please check the config to continue")
	}
}

func fatalIfErr(log *logrus.Logger, err error) {
	if err != nil {
		log.WithError(err).Fatal("can't continue")
	}
}

func getNginxStatus(log *logrus.Logger, config Config) string {
	nginxStatus := fmt.Sprintf("%v:%v/%v", config.NginxHost, config.NginxListenPort, config.NginxStatusURI)
	httpReq, err := http.NewRequest("GET", nginxStatus, bytes.NewBuffer([]byte("")))
	// http.NewRequest error
	fatalIfErr(log, err)
	code, data, err := runner.CallAPI(log, nil, httpReq, &http.Client{})
	if err != nil || code != 200 {
		log.WithFields(logrus.Fields{
			"code":                   code,
			"data":                   string(data),
			"httpReq":                httpReq,
			"config.NginxStatusPage": config.NginxHost,
			"config.NginxListenPort": config.NginxListenPort,
			"config.NginxStatusURI":  config.NginxStatusURI,
			"error":                  err,
		}).Fatal("Encountered error calling CallAPI")
		return ""
	}

	return string(data)
}

func scrapeStatus(log *logrus.Logger, status string) map[string]interface{} {

	multi := regexp.MustCompile(`Active connections: (\d+)`).FindString(status)
	contents := strings.Fields(multi)
	active := contents[2]

	multi = regexp.MustCompile(`(\d+)\s(\d+)\s(\d+)`).FindString(status)
	contents = strings.Fields(multi)
	accepts := contents[0]
	handled := contents[1]
	requests := contents[2]

	multi = regexp.MustCompile(`Reading: (\d+)`).FindString(status)
	contents = strings.Fields(multi)
	reading := contents[1]

	multi = regexp.MustCompile(`Writing: (\d+)`).FindString(status)
	contents = strings.Fields(multi)
	writing := contents[1]

	multi = regexp.MustCompile(`Waiting: (\d+)`).FindString(status)
	contents = strings.Fields(multi)
	waiting := contents[1]

	log.WithFields(logrus.Fields{
		"active":   active,
		"accepts":  accepts,
		"handled":  handled,
		"requests": requests,
		"reading":  reading,
		"writing":  writing,
		"waiting":  waiting,
	}).Debugf("Scraped NGINX values")
	return map[string]interface{}{
		"event_type":            "LoadBalancerSample",
		"provider":              PROVIDER,
		"nginx.net.connections": toInt(log, active),
		"nginx.net.accepts":     toInt(log, accepts),
		"nginx.net.handled":     toInt(log, handled),
		"nginx.net.requests":    toInt(log, requests),
		"nginx.net.writing":     toInt(log, writing),
		"nginx.net.waiting":     toInt(log, waiting),
		"nginx.net.reading":     toInt(log, reading),
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
