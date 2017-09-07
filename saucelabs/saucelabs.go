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

// NAME - the name of this thing
const NAME string = "saucelabs"

// PROVIDER - what app is sending the data
const PROVIDER string = "saucelabs"

// PROTOCOL_VERSION - nr-infra protocol version
const PROTOCOL_VERSION string = "1"

const url = "https://saucelabs.com/rest/v1/"

//SauceConfig is the keeper of the config
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

// User Metric holds the user metrics
type User struct {
	UserName string `json:"username"`
}

// Activity Metric tracks the users in activity
type Activity struct {
	SubAccounts map[string]SubAccount `json:"subaccounts"`
	Totals      SubAccount            `json:"totals"`
}

// SubAccount holds the job queued information
type SubAccount struct {
	InProgress int `json:"in progress"`
	All        int `json:"all"`
	Queued     int `json:"queued"`
}

// Data concurrency metric
type Data struct {
	Concurrency map[string]TeamData `json:"concurrency"`
}

type TeamData struct {
	Current   Allocation `json:"current"`
	Remaining Allocation `json:"remaining"`
}

type Allocation struct {
	Overall int `json:"overall"`
	Mac     int `json:"mac"`
	Manual  int `json:"manual"`
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

	var config = SauceConfig{
		SauceAPIUser: os.Getenv("SAUCE_API_USER"),
		SauceAPIKey:  os.Getenv("SAUCE_API_KEY"),
	}
	validateConfig(config)

	var metric = getMetric(log, config)
	data.Metrics = append(data.Metrics, metric...)

	fmt.Println("\n\n\n\n ")
	fatalIfErr(log, OutputJSON(data, prettyPrint))
}

func getMetric(log *logrus.Logger, config SauceConfig) []MetricData {
	client := http.Client{
		Timeout: time.Second * 20,
	}
	UserList := getUserList(client, config)
	UserActivity := getUserActivity(client, config)
	UserConcurrency := getConcurrency(client, config)

	fmt.Print("User List: ")
	fmt.Println(UserList)

	fmt.Print("\n\nUser Activity: ")
	fmt.Println(UserActivity)

	fmt.Print("\n\nUser Concurrency: ")
	fmt.Println(UserConcurrency)
	fmt.Println("\n\n ")

	var metricsData []MetricData

	// User List
	metricsData = append(metricsData, MetricData{
		"entity_name":           "SauceUserList",
		"event_type":            "SauceUserList",
		"provider":              "saucelabs",
		"userActivity.username": UserList,
	})

	// User Activity
	for key, value := range UserActivity.SubAccounts {
		metricsData = append(metricsData, MetricData{
			"entity_name":             "SauceUserActivity",
			"event_type":              "SauceUserActivity",
			"provider":                "saucelabs",
			"userActivity.username":   key,
			"userActivity.inprogress": value.InProgress,
			"userActivity.all":        value.All,
			"userActivity.queued":     value.Queued,
		})
	}
	metricsData = append(metricsData, MetricData{
		"entity_name":                   "SauceUserActivityTotal",
		"event_type":                    "SauceUserActivityTotal",
		"provider":                      "saucelabs",
		"userActivity.total.inprogress": UserActivity.Totals.InProgress,
		"userActivity.total.all":        UserActivity.Totals.All,
		"userActivity.total.queued":     UserActivity.Totals.Queued,
	})

	// User Cuncurency
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

	return metricsData
}

func validateConfig(config SauceConfig) {
	if config.SauceAPIUser == "" {
		log.Fatal("Config Yaml is missing SAUCE_API_USER value. Please check the config to continue")
	}
	if config.SauceAPIKey == "" {
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
		return nil
	}
	//set api key
	req.SetBasicAuth(config.SauceAPIUser, config.SauceAPIKey)
	//make request
	res, errdo := client.Do(req)
	if errdo != nil {
		return nil
	}
	body, errread := ioutil.ReadAll(res.Body)
	if errread != nil {
		return nil
	}
	err = json.Unmarshal(body, &userList)
	if err != nil {
		return nil
	}

	return userList
}

func getUserActivity(client http.Client, config SauceConfig) Activity {
	var userActivity Activity
	getUserActivityURL := url + config.SauceAPIUser + "/activity"

	//set url
	req, err := http.NewRequest(http.MethodGet, getUserActivityURL, nil)
	if err != nil {
		return Activity{}
	}
	//set api key
	req.SetBasicAuth(config.SauceAPIUser, config.SauceAPIKey)
	//make request
	res, errdo := client.Do(req)
	if errdo != nil {
		return Activity{}
	}
	body, errread := ioutil.ReadAll(res.Body)
	if errread != nil {
		return Activity{}
	}
	err = json.Unmarshal(body, &userActivity)
	if err != nil {
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
		return Data{}
	}
	//set api key
	req.SetBasicAuth(config.SauceAPIUser, config.SauceAPIKey)
	//make request
	res, errdo := client.Do(req)
	if errdo != nil {
		return Data{}
	}
	body, errread := ioutil.ReadAll(res.Body)
	if errread != nil {
		return Data{}
	}
	err = json.Unmarshal(body, &concurrencyList)
	if err != nil {
		return Data{}
	}
	return concurrencyList
}
