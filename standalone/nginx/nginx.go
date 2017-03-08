package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/GannettDigital/paas-api-utils/utilsHTTP"
	log "github.com/Sirupsen/logrus"
)

type NginxConfig struct {
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

func main() {
	// Setup the plugin's command line parameters
	verbose := flag.Bool("v", false, "Print more information to logs")
	pretty := flag.Bool("p", false, "Print pretty formatted JSON")
	flag.Parse()

	// Setup logging, redirect logs to stderr and configure the log level.
	log.SetOutput(os.Stderr)
	if *verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	// Initialize the output structure
	var data = PluginData{
		Name:            "nginx",
		ProtocolVersion: "1",
		PluginVersion:   "1.0.0",
		Status:          "OK",
		Inventory:       make(map[string]InventoryData),
		Metrics:         make([]MetricData, 0),
		Events:          make([]EventData, 0),
	}

	nginxListenPort := os.Getenv("NGINXLISTENPORT")
	nginxStatusURI := os.Getenv("NGINXSTATUSURI")
	nginxHost := os.Getenv("NGINXHOST")
	var nginxConf = NginxConfig{
		NginxListenPort: nginxListenPort,
		NginxHost:       nginxHost,
		NginxStatusURI:  nginxStatusURI,
	}

	var metric = scrapeStatus(getNginxStatus(nginxConf))

	data.Metrics = append(data.Metrics, metric)
	fatalIfErr(OutputJSON(data, *pretty))
}

func fatalIfErr(err error) {
	if err != nil {
		log.WithError(err).Fatal("can't continue")
	}
}

func getNginxStatus(config NginxConfig) string {
	nginxStatus := fmt.Sprintf("%v:%v/%v", config.NginxHost, config.NginxListenPort, config.NginxStatusURI)
	httpReq, err := http.NewRequest("GET", nginxStatus, bytes.NewBuffer([]byte("")))
	// http.NewRequest error
	if err != nil {
		log.WithFields(log.Fields{
			"nginxStatus": nginxStatus,
			"error":       err,
		}).Error("Encountered error creating http.NewRequest")
	}
	var runner utilsHTTP.HTTPRunner
	runner = utilsHTTP.HTTPRunnerImpl{}
	code, data, err := runner.CallAPI(log.New(), nil, httpReq, &http.Client{})
	if err != nil || code != 200 {
		log.WithFields(log.Fields{
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

func scrapeStatus(status string) map[string]interface{} {

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

	log.WithFields(log.Fields{
		"active":   active,
		"accepts":  accepts,
		"handled":  handled,
		"requests": requests,
		"reading":  reading,
		"writing":  writing,
		"waiting":  waiting,
	}).Info("Scraped NGINX values")
	return map[string]interface{}{
		"event_type":            "LoadBalancerSample",
		"provider":              "nginx",
		"nginx.net.connections": toInt(active),
		"nginx.net.accepts":     toInt(accepts),
		"nginx.net.handled":     toInt(handled),
		"nginx.net.requests":    toInt(requests),
		"nginx.net.writing":     toInt(writing),
		"nginx.net.waiting":     toInt(waiting),
		"nginx.net.reading":     toInt(reading),
	}
}

func toInt(value string) int {
	if value == "" {
		return 0
	} else {
		valueInt, err := strconv.Atoi(value)
		if err != nil {
			log.WithFields(log.Fields{
				"valueInt": valueInt,
				"error":    err,
			}).Error("Error converting value to int")

			return 0
		}

		return valueInt
	}
}
