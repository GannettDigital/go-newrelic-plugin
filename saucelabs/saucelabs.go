package saucelabs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
)

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

	var metric = getMetric(log, config)
	data.Metrics = append(data.Metrics, metric...)

	fatalIfErr(log, OutputJSON(data, prettyPrint))
}

func getMetric(log *logrus.Logger, config SauceConfig) []MetricData {
	var metricsData []MetricData

	client := http.Client{
		Timeout: time.Second * 20,
	}
	UserList := getUserList(client, config)
	UserActivity := getUserActivity(client, config)
	UserConcurrency := getConcurrency(client, config)
	UserHistory := getUsage(client, config)

	// User List Metrics
	metricsData = append(metricsData, MetricData{
		"entity_name":           "SauceUserList",
		"event_type":            "SauceUserList",
		"provider":              "saucelabs",
		"userActivity.username": UserList,
	})

	// User Activity Metrics
	for key, value := range UserActivity.SubAccounts {
		metricsData = append(metricsData, MetricData{
			"entity_name":             "SauceUserActivity",
			"event_type":              "SauceUserActivity",
			"provider":                "saucelabs",
			"userActivity.username":   key,
			"userActivity.inProgress": value.InProgress,
			"userActivity.all":        value.All,
			"userActivity.queued":     value.Queued,
		})
	}
	metricsData = append(metricsData, MetricData{
		"entity_name":                   "SauceUserActivityTotal",
		"event_type":                    "SauceUserActivityTotal",
		"provider":                      "saucelabs",
		"userActivity.total.inProgress": UserActivity.Totals.InProgress,
		"userActivity.total.all":        UserActivity.Totals.All,
		"userActivity.total.queued":     UserActivity.Totals.Queued,
	})

	// User Concurency Metrics
	for key, value := range UserConcurrency.Concurrency {
		metricsData = append(metricsData, MetricData{
			"entity_name":                      "SauceUserConcurrency",
			"event_type":                       "SauceUserConcurrency",
			"provider":                         "saucelabs",
			"userConcurrency.username":         key,
			"userConcurrency.current.overall":  value.Current.Overall,
			"userConcurrency.current.mac":      value.Current.Mac,
			"userConcurrency.current.manual":   value.Current.Manual,
			"userConcurrency.Remaning.overall": value.Remaining.Overall,
			"userConcurrency.Remaning.mac":     value.Remaining.Mac,
			"userConcurrency.Remaning.manual":  value.Remaining.Manual,
		})
	}

	// User Usage
	for i := range UserHistory.Usage {
		metricsData = append(metricsData, MetricData{
			"entity_name":                 "SauceUserHistory",
			"event_type":                  "SauceUserHistory",
			"provider":                    "saucelabs",
			"userHistory.username":        UserHistory.UserName,
			"userHistory.date":            UserHistory.Usage[i][0].(string),
			"userHistory.totalJobs":       UserHistory.Usage[i][1].([]interface{})[0].(float64),
			"userHistory.totalTimeInSecs": UserHistory.Usage[i][1].([]interface{})[1].(float64),
		})
	}
	return metricsData
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

func fatalIfErr(log *logrus.Logger, err error) {
	if err != nil {
		log.WithError(err).Fatal("can't continue")
	}
}

func getUserList(client http.Client, config SauceConfig) []User {
	var userList []User
	getUserListURL := url + "users/" + config.SauceAPIUser + "/subaccounts"

	//set url
	req, err := http.NewRequest(http.MethodGet, getUserListURL, nil)
	if err != nil {
		log.Fatal("Bad New Request")
		return userList
	}
	//set api key
	req.SetBasicAuth(config.SauceAPIUser, config.SauceAPIKey)
	//make request
	res, errdo := client.Do(req)
	if errdo != nil {
		log.Fatal("Bad Client Request")
		return userList
	}
	body, errread := ioutil.ReadAll(res.Body)
	if errread != nil {
		log.Fatal("Error Reading Body")
		return userList
	}
	err = json.Unmarshal(body, &userList)
	if err != nil {
		log.Fatal("Error Unmarshalling Body")
		return userList
	}
	return userList
}

func getUserActivity(client http.Client, config SauceConfig) Activity {
	var userActivity Activity
	getUserActivityURL := url + config.SauceAPIUser + "/activity"

	//set url
	req, err := http.NewRequest(http.MethodGet, getUserActivityURL, nil)
	if err != nil {
		log.Fatal("Bad New Request")
		return Activity{}
	}
	//set api key
	req.SetBasicAuth(config.SauceAPIUser, config.SauceAPIKey)
	//make request
	res, errdo := client.Do(req)
	if errdo != nil {
		log.Fatal("Bad Client Request")
		return Activity{}
	}
	body, errread := ioutil.ReadAll(res.Body)
	if errread != nil {
		log.Fatal("Error Reading Body")
		return Activity{}
	}
	err = json.Unmarshal(body, &userActivity)
	if err != nil {
		log.Fatal("Error Unmarshalling Body")
		return Activity{}
	}
	return userActivity
}

func getConcurrency(client http.Client, config SauceConfig) Data {
	var concurrencyList Data
	getConcurrencyURL := url + "users/" + config.SauceAPIUser + "/concurrency"

	//set url
	req, err := http.NewRequest(http.MethodGet, getConcurrencyURL, nil)
	if err != nil {
		log.Fatal("Bad New Request")
		return Data{}
	}
	//set api key
	req.SetBasicAuth(config.SauceAPIUser, config.SauceAPIKey)
	//make request
	res, errdo := client.Do(req)
	if errdo != nil {
		log.Fatal("Bad Client Request")
		return Data{}
	}
	body, errread := ioutil.ReadAll(res.Body)
	if errread != nil {
		log.Fatal("Error Reading Body")
		return Data{}
	}
	err = json.Unmarshal(body, &concurrencyList)
	if err != nil {
		log.Fatal("Error Unmarshalling Body")
		return Data{}
	}
	return concurrencyList
}

func getUsage(client http.Client, config SauceConfig) History {

	fmt.Println("\n\nTESTLINE\n\n ")

	var usageList History
	getUsageURL := url + "users/" + config.SauceAPIUser + "/usage"

	//set url
	req, err := http.NewRequest(http.MethodGet, getUsageURL, nil)
	if err != nil {
		log.Fatal("Bad New Request")
		return History{}
	}
	//set api key
	req.SetBasicAuth(config.SauceAPIUser, config.SauceAPIKey)
	//make request
	res, errdo := client.Do(req)
	if errdo != nil {
		log.Fatal("Bad Client Request")
		return History{}
	}
	body, errread := ioutil.ReadAll(res.Body)
	if errread != nil {
		log.Fatal("Error Reading Body")
		return History{}
	}

	err = json.Unmarshal(body, &usageList)
	if err != nil {
		log.Fatal("Error Unmarshalling Body")
		return History{}
	}
	return usageList
}
