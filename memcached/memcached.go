package memcached

import (
	"bufio"
	"fmt"
	"github.com/GannettDigital/go-newrelic-plugin/helpers"
	"github.com/Sirupsen/logrus"
	"net"
	"os"
	"strings"
)

const NAME string = "memcached"
const PROVIDER string = "memcached"
const PROTOCOL_VERSION string = "1"
const PLUGIN_VERSION string = "1.0.0"
const STATUS string = "OK"

//MemcachedConfig is the keeper of the config
type MemcachedConfig struct {
	MemcachedHost string
	MemcachedPort string
	Commands      string
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
	Status          string                   `json:"status"`
	Metrics         []MetricData             `json:"metrics"`
	Inventory       map[string]InventoryData `json:"inventory"`
	Events          []EventData              `json:"events"`
}

var localLog *logrus.Logger

func Run(log *logrus.Logger, prettyPrint bool, version string) {
	// Initialize the output structure
	localLog = log
	var data = PluginData{
		Name:            NAME,
		PluginVersion:   PLUGIN_VERSION,
		ProtocolVersion: PROTOCOL_VERSION,
		Status:          STATUS,
		Metrics:         make([]MetricData, 0),
		Inventory:       make(map[string]InventoryData),
		Events:          make([]EventData, 0),
	}

	var config = MemcachedConfig{
		MemcachedHost: os.Getenv("MEMCACHED_HOST"),
		MemcachedPort: os.Getenv("MEMCACHED_PORT"),
		Commands:      os.Getenv("COMMANDS"),
	}
	validateConfig(config)

	metric, err := getMetric(config)
	if err != nil {
		data.Status = err.Error()
	}
	data.Metrics = append(data.Metrics, metric)
	fatalIfErr(helpers.OutputJSON(data, prettyPrint), "OutputJSON error")
}

func getMetric(config MemcachedConfig) (map[string]interface{}, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", config.MemcachedHost, config.MemcachedPort))
	if err != nil {
		localLog.WithError(err).Error(fmt.Sprintf("getMetric: Cannot connect to memcached %s:%s", config.MemcachedHost, config.MemcachedPort))
		return nil, err
	}

	metrics := map[string]interface{}{
		"event_type": "DatastoreSample",
		"provider":   PROVIDER,
	}

	for _, command := range strings.Split(config.Commands, ",") {
		command = strings.TrimSpace(command)
		localLog.Debug(fmt.Sprintf("scanResult: command: %s", command))
		fmt.Fprintf(conn, "%s\r\n", command)
		scanner := bufio.NewScanner(bufio.NewReader(conn))
		scanResult(scanner, command, metrics)
	}
	return metrics, nil
}

func scanResult(scanner *bufio.Scanner, command string, metrics map[string]interface{}) {
	for scanner.Scan() {
		localLog.Debug(fmt.Sprintf("scanResult: scanning..."))

		if err := scanner.Err(); err != nil {
			localLog.WithError(err).Error("reading scanning connection")
		}
		result := strings.TrimSuffix(scanner.Text(), "\r")
		localLog.Debug(fmt.Sprintf("scanResult: result: %s", result))
		if strings.Compare("END", result) == 0 {
			break
		}
		line := strings.Split(result, " ")

		name := metricName(command, line[1])
		metrics[name] = helpers.AsValue(strings.Join(line[2:], " "))
	}
}

func metricName(command string, metric string) string {
	localLog.Debug(fmt.Sprintf("metricName: command: %s metric: %s", command, metric))
	line := strings.Split(command, " ")
	result := "memcached"
	localLog.Debug(fmt.Sprintf("metricName: result1: %s", result))
	if len(line) == 2 {
		result = fmt.Sprintf("%s.%s", result, (strings.Split(command, " "))[1])
		localLog.Debug(fmt.Sprintf("metricName: result2: %s", result))
	}
	result = fmt.Sprintf("%s.%s", result, helpers.CamelCase(metric))
	localLog.Debug(fmt.Sprintf("metricName: result3: %s", result))
	return result
}

func validateConfig(config MemcachedConfig) {
	if config.MemcachedHost == "" {
		localLog.Fatal("Config Yaml is missing MEMCACHED_HOST value. Please check the config to continue")
	}
	if config.MemcachedPort == "" {
		localLog.Fatal("Config Yaml is missing MEMCACHED_PORT value. Please check the config to continue")
	}
	if len(config.Commands) < 1 {
		localLog.Fatal("Config Yaml is missing COMMANDS value. Please check the config to continue")
	}
}

func fatalIfErr(err error, msg string) {
	if err != nil {
		localLog.WithError(err).Fatal(msg)
	}
}
