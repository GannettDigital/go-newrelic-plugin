package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
)

var log *logrus.Logger

const EVENT_TYPE string = "DataStoreSample"
const NAME string = "couchbase"
const PROVIDER string = "couchbase" //we might want to make this an env tied to nginx version or app name maybe...
const VERSION string = "1.0.0"
const PROTOCOL_VERSION string = "1"

//CouchbaseConfig is the keeper of the config
type CouchbaseConfig struct {
	CouchbaseHost string
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
	log = logrus.New()
}

func main() {
	// Setup the plugin's command line parameters
	verbose := flag.Bool("v", false, "Print more information to logs")
	pretty := flag.Bool("p", false, "Print pretty formatted JSON")
	version := flag.Bool("version", false, "Print the version and exit")
	flag.Parse()

	if *version {
		fmt.Println(VERSION)
		os.Exit(1)
	}

	// Setup logging, redirect logs to stderr and configure the log level.
	log.Out = os.Stderr
	if *verbose {
		log.Level = logrus.DebugLevel
	} else {
		log.Level = logrus.InfoLevel
	}

	// Initialize the output structure
	var data = PluginData{
		Name:            NAME,
		ProtocolVersion: PROTOCOL_VERSION,
		PluginVersion:   VERSION,
		Inventory:       make(map[string]InventoryData),
		Metrics:         make([]MetricData, 0),
		Events:          make([]EventData, 0),
	}

	var config = CouchbaseConfig{
		CouchbaseHost: os.Getenv("KEY"),
	}
	validateConfig(config)

	var metric = getMetric(config)

	data.Metrics = append(data.Metrics, metric)
	fatalIfErr(OutputJSON(data, *pretty))
}

func getMetric(config CouchbaseConfig) map[string]interface{} {
	return map[string]interface{}{
		"event_type":     "LoadBalancerSample",
		"provider":       PROVIDER,
		"couchbase.stat": 1,
	}
}

func validateConfig(config CouchbaseConfig) {
	if config.CouchbaseHost == "" {
		log.Fatal("Config Yaml is missing values. Please check the config to continue")
	}
}

func fatalIfErr(err error) {
	if err != nil {
		log.WithError(err).Fatal("can't continue")
	}
}
