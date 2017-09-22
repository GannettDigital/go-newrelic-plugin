package saucelabs

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/Sirupsen/logrus"
)

// Name - the name of this thing
const Name string = "saucelabs"

// Provider - what app is sending the data
const Provider string = "saucelabs"

// ProtocolVersion - nr-infra protocol version
const ProtocolVersion string = "1"

// SauceConfig is the keeper of the config
type SauceConfig struct {
	SauceAPIUser string
	SauceAPIKey  string
}

// SauceClient holds sauce http connection information
type SauceClient struct {
	URL    url.URL
	Client *http.Client
	Config SauceConfig
}

// NewSauceClient creates a new sauce client
func NewSauceClient(config SauceConfig) (*SauceClient, error) {
	base, err := url.Parse("https://saucelabs.com/rest/v1/")
	if err != nil {
		return nil, err
	}
	return &SauceClient{
		*base,
		new(http.Client),
		config,
	}, nil
}

func (sc *SauceClient) do(method, path string, into interface{}, args map[string]string) error {
	baseURL := sc.URL
	request := &http.Request{
		Method: method,
		URL:    &baseURL,
		Header: http.Header{
			"Accept": {"application/json"},
		},
	}
	request.URL.Path += path
	request.SetBasicAuth(sc.Config.SauceAPIUser, sc.Config.SauceAPIKey)

	response, responseErr := sc.Client.Do(request)
	if responseErr != nil {
		return responseErr
	}
	defer response.Body.Close()

	decodeErr := json.NewDecoder(response.Body).Decode(into)
	if decodeErr != nil {
		return decodeErr
	}

	return nil
}

//GetUserList retrieves all the subaccounts of a parent account
func (sc *SauceClient) GetUserList() ([]User, error) {
	var response []User
	getUserListURL := fmt.Sprintf("users/%v/subaccounts", sc.Config.SauceAPIUser)

	err := sc.do(http.MethodGet, getUserListURL, &response, nil)
	if err != nil {
		return []User{}, err
	}

	return response, nil
}

//GetUserActivity retrieves the activity of an account or all child accounts
func (sc *SauceClient) GetUserActivity() (Activity, error) {
	var response Activity
	getUserActivityURL := fmt.Sprintf("users/%v/activity", sc.Config.SauceAPIUser)

	err := sc.do(http.MethodGet, getUserActivityURL, &response, nil)
	if err != nil {
		return Activity{}, err
	}

	return response, nil
}

//GetConcurrency retrieves the concurrency for an account or all child accounts
func (sc *SauceClient) GetConcurrency() (Data, error) {
	var response Data
	getConcurrencyURL := fmt.Sprintf("users/%v/concurrency", sc.Config.SauceAPIUser)

	err := sc.do(http.MethodGet, getConcurrencyURL, &response, nil)
	if err != nil {
		return Data{}, err
	}

	return response, nil
}

//GetUsage retrieves the usage metric for the passed account
func (sc *SauceClient) GetUsage() (History, error) {
	var response History
	getUsageURL := fmt.Sprintf("users/%v/usage", sc.Config.SauceAPIUser)

	err := sc.do(http.MethodGet, getUsageURL, &response, nil)
	if err != nil {
		return History{}, err
	}

	return response, nil
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

	sc, scErr := NewSauceClient(config)
	if scErr != nil {
		log.WithError(scErr).Error("Error creating saucelabs client")
		return
	}

	metric, metricsErr := getMetric(log, config, sc)
	if metricsErr != nil {
		log.WithError(metricsErr).Error("Error collecting metrics")
		return
	}

	data.Metrics = append(data.Metrics, metric...)

	fatalIfErr(log, OutputJSON(data, prettyPrint))
}

func fatalIfErr(log *logrus.Logger, err error) {
	if err != nil {
		log.WithError(err).Fatal("Can't Continue")
	}
}

func getMetric(log *logrus.Logger, config SauceConfig, sc *SauceClient) ([]MetricData, error) {
	var metricsData []MetricData

	userList, userListErr := sc.GetUserList()
	if userListErr != nil {
		log.WithError(userListErr).Error("Error collecting user list metrics")
		return nil, userListErr
	}
	userActivity, userActivityErr := sc.GetUserActivity()
	if userActivityErr != nil {
		log.WithError(userActivityErr).Error("Error collecting user activty metrics")
		return nil, userActivityErr
	}
	userConcurrency, userConcurrencyErr := sc.GetConcurrency()
	if userConcurrencyErr != nil {
		log.WithError(userConcurrencyErr).Error("Error collecting user concurrency metrics")
		return nil, userConcurrencyErr
	}
	userHistory, userHistoryErr := sc.GetUsage()
	if userHistoryErr != nil {
		log.WithError(userHistoryErr).Error("Error collecting user usage metrics")
		return nil, userHistoryErr
	}

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
	return metricsData, nil
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
