package saucelabs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/GannettDigital/paas-api-utils/utilsHTTP"
)
var runner utilsHTTP.HTTPRunner

// Name - the name of this thing
const Name string = "saucelabs"

// Provider - what app is sending the data
const Provider string = "saucelabs"

// ProtocolVersion - nr-infra protocol version
const ProtocolVersion string = "1"

const url = "https://saucelabs.com/rest/v1/"

// SauceConfig is the keeper of the config
type SauceConfig struct {
	SauceAPIUser string
	SauceAPIKey  string
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

// User Metric holds the usernames for getUserList
type User struct {
	UserName string `json:"username"`
}

// Activity Metric tracks the users in activity
type Activity struct {
	SubAccounts map[string]SubAccount `json:"subaccounts"`
	Totals      SubAccount            `json:"totals"`
}

// SubAccount holds the active job queued information
type SubAccount struct {
	InProgress int `json:"in progress"`
	All        int `json:"all"`
	Queued     int `json:"queued"`
}

// Data concurrency metric
type Data struct {
	Concurrency map[string]TeamData `json:"concurrency"`
}

// TeamData for concurrency
type TeamData struct {
	Current   Allocation `json:"current"`
	Remaining Allocation `json:"remaining"`
}

// Allocation holds TeamData concurrency stats
type Allocation struct {
	Overall int `json:"overall"`
	Mac     int `json:"mac"`
	Manual  int `json:"manual"`
}

// History holds the username and total number of jobs and VM time used, in seconds grouped by day.
type History struct {
	UserName string          `json:"username"`
	Usage    [][]interface{} `json:"usage"`
}

var client = http.Client{
	Timeout: time.Second * 20,
}

// OutputJSON takes an object and prints it as a JSON string to the stdout.
// If the pretty attribute is set to true, the JSON will be indented for easy reading.
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

// Run - Function that is ran from the main cmd
func Run(log *logrus.Logger, prettyPrint bool, version string) {

	// Initialize the output structure
	var data = PluginData{
		Name:            Name,
		ProtocolVersion: ProtocolVersion,
		PluginVersion:   version,
		Inventory:       make(map[string]InventoryData),
		Metrics:         make([]MetricData, 0),
		Events:          make([]EventData, 0),
	}

	var config = SauceConfig{
		SauceAPIUser: os.Getenv("SAUCE_API_USER"),
		SauceAPIKey:  os.Getenv("SAUCE_API_KEY"),
	}
	validateConfig(config)

	var metric = getMetric(log, config, client)
	data.Metrics = append(data.Metrics, metric...)

	fatalIfErr(log, OutputJSON(data, prettyPrint))
}

func fatalIfErr(log *logrus.Logger, err error) {
	if err != nil {
		log.WithError(err).Fatal("Can't Continue")
	}
}

func getMetric(log *logrus.Logger, config SauceConfig, client http.Client) []MetricData {
	var metricsData []MetricData

	userList := getUserList(client, config)
	userActivity := getUserActivity(client, config)
	userConcurrency := getConcurrency(client, config)
	userHistory := getUsage(client, config)

	// User List Metrics
	metricsData = append(metricsData, MetricData{
		"entity_name":                     "SauceUserList",
		"event_type":                      "SauceUserList",
		"provider":                        "saucelabs",
		"saucelabs.userActivity.username": userList,
	})

	// User Activity Metrics
	for key, value := range userActivity.SubAccounts {
		metricsData = append(metricsData, MetricData{
			"entity_name":                       "SauceUserActivity",
			"event_type":                        "SauceUserActivity",
			"provider":                          "saucelabs",
			"saucelabs.userActivity.username":   key,
			"saucelabs.userActivity.inProgress": value.InProgress,
			"saucelabs.userActivity.all":        value.All,
			"saucelabs.userActivity.queued":     value.Queued,
		})
	}
	metricsData = append(metricsData, MetricData{
		"entity_name": "SauceUserActivityTotal",
		"event_type":  "SauceUserActivityTotal",
		"provider":    "saucelabs",
		"saucelabs.userActivity.total.inProgress": userActivity.Totals.InProgress,
		"saucelabs.userActivity.total.all":        userActivity.Totals.All,
		"saucelabs.userActivity.total.queued":     userActivity.Totals.Queued,
	})

	// User Concurency Metrics
	for key, value := range userConcurrency.Concurrency {
		metricsData = append(metricsData, MetricData{
			"entity_name":                                "SauceUserConcurrency",
			"event_type":                                 "SauceUserConcurrency",
			"provider":                                   "saucelabs",
			"saucelabs.userConcurrency.username":         key,
			"saucelabs.userConcurrency.current.overall":  value.Current.Overall,
			"saucelabs.userConcurrency.current.mac":      value.Current.Mac,
			"saucelabs.userConcurrency.current.manual":   value.Current.Manual,
			"saucelabs.userConcurrency.Remaning.overall": value.Remaining.Overall,
			"saucelabs.userConcurrency.Remaning.mac":     value.Remaining.Mac,
			"saucelabs.userConcurrency.Remaning.manual":  value.Remaining.Manual,
		})
	}

	// User Usage
	for index := range userHistory.Usage {
		metricsData = append(metricsData, MetricData{
			"entity_name":                           "SauceUserHistory",
			"event_type":                            "SauceUserHistory",
			"provider":                              "saucelabs",
			"saucelabs.userHistory.username":        userHistory.UserName,
			"saucelabs.userHistory.date":            getHistoryDate(userHistory, index),
			"saucelabs.userHistory.totalJobs":       getHistoryTotalJobs(userHistory, index),
			"saucelabs.userHistory.totalTimeInSecs": getHistoryTotalTime(userHistory, index),
		})
	}
	return metricsData
}
func getHistoryDate(userHistory History, index int) string {
	r, _ := regexp.Compile("([0-9]{4})+[-]([0-9]{1,2})+[-]+([0-9]{1,2})")
	if r.MatchString(userHistory.Usage[index][0].(string)) {
		return userHistory.Usage[index][0].(string)
	}
	log.Fatal("Error parsing users history date")
	return ""
}
func getHistoryTotalJobs(userHistory History, index int) float64 {
	totalJobs, check := userHistory.Usage[index][1].([]interface{})[0].(float64)
	if check == true {
		return totalJobs
	}
	log.Fatal("Error parsing users total jobs")
	return 0
}
func getHistoryTotalTime(userHistory History, index int) float64 {
	totalTime, check := userHistory.Usage[index][1].([]interface{})[1].(float64)
	if check == true {
		return totalTime
	}
	log.Fatal("Error parsing users total time")
	return 0
}

func validateConfig(config SauceConfig) {
	if config.SauceAPIUser == "" && config.SauceAPIKey == "" {
		log.Fatal("Config Yaml is missing SAUCE_API_USER and SAUCE_API_KEY values. Please check the config to continue")
	} else if config.SauceAPIUser == "" {
		log.Fatal("Config Yaml is missing SAUCE_API_USER value. Please check the config to continue")
	} else if config.SauceAPIKey == "" {
		log.Fatal("Config Yaml is missing SAUCE_API_KEY value. Please check the config to continue")
	}
}

func getUserList(client http.Client, config SauceConfig) []User {
	var userList []User
	getUserListURL := url + "users/" + config.SauceAPIUser + "/subaccounts"

	body := httpRequest(client, config, getUserListURL)

	var err = json.Unmarshal(body, &userList)
	if err != nil {
		log.Fatal("Error Unmarshalling Body")
		return userList
	}
	return userList
}

func getUserActivity(client http.Client, config SauceConfig) Activity {
	var userActivity Activity
	getUserActivityURL := url + config.SauceAPIUser + "/activity"

	body := httpRequest(client, config, getUserActivityURL)

	var err = json.Unmarshal(body, &userActivity)
	if err != nil {
		log.Fatal("Error Unmarshalling Body")
		return Activity{}
	}
	return userActivity
}

func getConcurrency(client http.Client, config SauceConfig) Data {
	var concurrencyList Data
	getConcurrencyURL := url + "users/" + config.SauceAPIUser + "/concurrency"

	body := httpRequest(client, config, getConcurrencyURL)

	var err = json.Unmarshal(body, &concurrencyList)
	if err != nil {
		log.Fatal("Error Unmarshalling Body")
		return Data{}
	}
	return concurrencyList
}

func getUsage(client http.Client, config SauceConfig) History {
	var usageList History
	getUsageURL := url + "users/" + config.SauceAPIUser + "/usage"

	body := httpRequest(client, config, getUsageURL)

	var err = json.Unmarshal(body, &usageList)
	if err != nil {
		log.Fatal("Error Unmarshalling Body")
		return History{}
	}
	return usageList
}

func httpRequest(client http.Client, config SauceConfig, url string) []byte {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal("Bad New Request")
		return nil
	}
	//set api key
	req.SetBasicAuth(config.SauceAPIUser, config.SauceAPIKey)
	//make request
	res, errdo := client.Do(req)
	if errdo != nil {
		log.Fatal("Bad Client Request")
		return nil
	}
	body, errread := ioutil.ReadAll(res.Body)
	if errread != nil {
		log.Fatal("Error Reading Body")
		return nil
	}
	return body
}
