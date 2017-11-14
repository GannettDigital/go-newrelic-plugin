package saucelabs

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
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

func (sc *SauceClient) do(method string, path Path, into interface{}, args map[string]string) error {
	baseURL := sc.URL
	request := &http.Request{
		Method: method,
		URL:    &baseURL,
		Header: http.Header{
			"Accept": {"application/json"},
		},
	}

	request.URL.Path += path.Path
	if path.Parameter != nil {
		parameters := url.Values{}
		for index := range path.Parameter {
			parameters.Add(path.Parameter[index].key, path.Parameter[index].value)
			request.URL.RawQuery = parameters.Encode()
		}
		request.URL.RawQuery = parameters.Encode()
	}

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
	pathURL := Path{Path: getUserListURL}

	err := sc.do(http.MethodGet, pathURL, &response, nil)
	if err != nil {
		return []User{}, err
	}
	return response, nil
}

//GetUserActivity retrieves the activity of an account or all child accounts
func (sc *SauceClient) GetUserActivity() (Activity, error) {
	var response Activity
	getUserActivityURL := fmt.Sprintf("users/%v/activity", sc.Config.SauceAPIUser)
	pathURL := Path{Path: getUserActivityURL}

	err := sc.do(http.MethodGet, pathURL, &response, nil)
	if err != nil {
		return Activity{}, err
	}
	return response, nil
}

//GetConcurrency retrieves the concurrency for an account or all child accounts
func (sc *SauceClient) GetConcurrency() (Data, error) {
	var response Data
	getConcurrencyURL := fmt.Sprintf("users/%v/concurrency", sc.Config.SauceAPIUser)
	pathURL := Path{Path: getConcurrencyURL}

	err := sc.do(http.MethodGet, pathURL, &response, nil)
	if err != nil {
		return Data{}, err
	}
	return response, nil
}

//GetUsage retrieves the usage metric for the passed account
func (sc *SauceClient) GetUsage() (HistoryFormated, error) {
	var response History
	getUsageURL := fmt.Sprintf("users/%v/usage", sc.Config.SauceAPIUser)
	pathURL := Path{Path: getUsageURL}

	err := sc.do(http.MethodGet, pathURL, &response, nil)
	if err != nil {
		return HistoryFormated{}, err
	}

	var formatedResponse HistoryFormated
	formatedResponse.UserName = response.UserName
	for index := range response.Usage {
		var testInfo TestInfo
		testInfo.Executed = getHistoryTotalJobs(response, index)
		testInfo.Time = getHistoryTotalTime(response, index)

		var usageList UsageList
		usageList.Date = getHistoryDate(response, index)
		usageList.testInfoList = testInfo
		formatedResponse.Usage = append(formatedResponse.Usage, usageList)
	}
	return formatedResponse, nil
}

// GetErrors retrieves the error metrics for the passed account
func (sc *SauceClient) GetErrors(startDateString string, endDateString string) (Errors, error) {
	var response Errors

	path := "analytics/trends/errors"
	pathURL := getPathURL(startDateString, endDateString, path)

	err := sc.do(http.MethodGet, pathURL, &response, nil)
	if err != nil {
		return Errors{}, err
	}

	return response, nil
}

// Trends holds data for GetBuildTrends
type Trends struct {
	Builds BuildItems `json:"builds"`
}

// BuildItems holds data for GetBuildTrends
type BuildItems struct {
	BuildItems []Items `json:"items"`
}

// Items holds data for GetBuildTrends
type Items struct {
	BuildName        string      `json:"name"`
	OwnerName        string      `json:"owner"`
	TestsCount       int         `json:"tests_count"`
	Duration         int         `json:"duration"`
	DurationAbsolute int         `json:"duration_absolute"`
	DurationTestMax  int         `json:"duration_test_max"`
	StartTime        string      `json:"start_time"`
	EndTime          string      `json:"end_time"`
	ItemsList        []ItemsList `json:"tests"`
}

func (sc *SauceClient) GetBuildTrends(startDateString string, endDateString string) (Trends, error) {
	var response Trends
	path := "analytics/trends/builds_tests"
	pathURL := getPathURL(startDateString, endDateString, path)

	err := sc.do(http.MethodGet, pathURL, &response, nil)
	if err != nil {
		return Trends{}, err
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

// Path - Holds the path for url
type Path struct {
	Path      string
	Parameter []Parameter
}

// Parameter - Holds the Parameters information for URL
type Parameter struct {
	key   string
	value string
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

// HistoryFormated holds the username and total number of jobs and VM time used, in seconds grouped by day formated for testing.
type HistoryFormated struct {
	UserName string      `json:"username"`
	Usage    []UsageList `json:"usage"`
}

// UsageList holds the Date and a testInfo list for a particular usage object
type UsageList struct {
	Date         time.Time
	testInfoList TestInfo
}

// TestInfo holds the Time executed and the duration executed
type TestInfo struct {
	Executed float64
	Time     float64
}

// Errors - holds the buckets of errors
type Errors struct {
	Buckets []BucketsList `json:"buckets"`
}

// BucketsList - used in errors struct
type BucketsList struct {
	Name  string      `json:"name"`
	Count int         `json:"count"`
	Items []ItemsList `json:"items"`
}

// ItemsList - used in bucketslist and errors struct
type ItemsList struct {
	ID           string `json:"id"`
	Owner        string `json:"owner"`
	Name         string `json:"name"`
	Build        string `json:"build"`
	CreationTime string `json:"creation_time"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	Duration     int    `json:"duration"`
	Status       string `json:"status"`
	Error        string `json:"error"`
	OS           string `json:"os"`
	Browser      string `json:"browser"`
	DetailsURL   string `json:"details_url"`
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

	metric, metricsErr := getMetrics(log, config, sc)
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

func getMetrics(log *logrus.Logger, config SauceConfig, sc *SauceClient) ([]MetricData, error) {
	var metricsData []MetricData

	userList, userListErr := sc.GetUserList()
	if userListErr != nil {
		log.WithError(userListErr).Error("Error collecting user list metrics")
		return nil, userListErr
	}
	userActivity, userActivityErr := sc.GetUserActivity()
	if userActivityErr != nil {
		log.WithError(userActivityErr).Error("Error collecting user activity metrics")
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

	endDate := time.Now()
	endDateString := endDate.Format(time.RFC3339)[:19]
	startDateSecs := endDate.Unix() - 2505600
	startDateString := time.Unix(startDateSecs, 0).Format(time.RFC3339)[:19]
	errorHistory, errorHistoryErr := sc.GetErrors(startDateString, endDateString)
	if errorHistoryErr != nil {
		log.WithError(errorHistoryErr).Error("Error collecting error metrics")
		return nil, userHistoryErr
	}
	trendsHistory, errorTrendsHistory := sc.GetBuildTrends(startDateString, endDateString)
	if errorTrendsHistory != nil {
		log.WithError(errorTrendsHistory).Error("Error collecting build trends metrics")
		return nil, errorTrendsHistory
	}

	// User List Metrics
	for index := range userList {
		metricsData = append(metricsData, MetricData{
			"entity_name":        "SauceLabs",
			"event_type":         "SauceLabs",
			"provider":           "saucelabs",
			"saucelabs.username": userList[index].UserName,
		})
	}

	// User Activity Metrics
	for key, value := range userActivity.SubAccounts {
		metricsData = append(metricsData, MetricData{
			"entity_name":                       "SauceLabs",
			"event_type":                        "SauceLabs",
			"provider":                          "saucelabs",
			"saucelabs.username":                key,
			"saucelabs.userActivity.inProgress": value.InProgress,
			"saucelabs.userActivity.all":        value.All,
			"saucelabs.userActivity.queued":     value.Queued,
		})
	}
	metricsData = append(metricsData, MetricData{
		"entity_name": "SauceLabs",
		"event_type":  "SauceLabs",
		"provider":    "saucelabs",
		"saucelabs.userActivity.total.inProgress": userActivity.Totals.InProgress,
		"saucelabs.userActivity.total.all":        userActivity.Totals.All,
		"saucelabs.userActivity.total.queued":     userActivity.Totals.Queued,
	})

	// User Concurency Metrics
	for key, value := range userConcurrency.Concurrency {
		metricsData = append(metricsData, MetricData{
			"entity_name":                                "SauceLabs",
			"event_type":                                 "SauceLabs",
			"provider":                                   "saucelabs",
			"saucelabs.username":                         key,
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
			"entity_name":                           "SauceLabs",
			"event_type":                            "SauceLabs",
			"provider":                              "saucelabs",
			"saucelabs.username":                    userHistory.UserName,
			"saucelabs.userHistory.date":            userHistory.Usage[index].Date,
			"saucelabs.userHistory.totalJobs":       userHistory.Usage[index].testInfoList.Executed,
			"saucelabs.userHistory.totalTimeInSecs": userHistory.Usage[index].testInfoList.Time,
		})
	}

	// Error History
	for i := range errorHistory.Buckets {
		for j := range errorHistory.Buckets[i].Items {
			metricsData = append(metricsData, MetricData{
				"entity_name":                       "SauceLabs",
				"event_type":                        "SauceLabs",
				"provider":                          "saucelabs",
				"saucelabs.name":                    errorHistory.Buckets[i].Name,
				"saucelabs.userError.count":         errorHistory.Buckets[i].Count,
				"saucelabs.userError.id":            errorHistory.Buckets[i].Items[j].ID,
				"saucelabs.username":                errorHistory.Buckets[i].Items[j].Owner,
				"saucelabs.userError.build":         errorHistory.Buckets[i].Items[j].Build,
				"saucelabs.userError.creation_time": errorHistory.Buckets[i].Items[j].CreationTime,
				"saucelabs.userError.start_time":    errorHistory.Buckets[i].Items[j].StartTime,
				"saucelabs.userError.end_time":      errorHistory.Buckets[i].Items[j].EndTime,
				"saucelabs.userError.duration":      errorHistory.Buckets[i].Items[j].Duration,
				"saucelabs.userError.status":        errorHistory.Buckets[i].Items[j].Status,
				"saucelabs.userError.error":         errorHistory.Buckets[i].Items[j].Error,
				"saucelabs.userError.os":            errorHistory.Buckets[i].Items[j].OS,
				"saucelabs.userError.browser":       errorHistory.Buckets[i].Items[j].Browser,
				"saucelabs.userError.details_url":   errorHistory.Buckets[i].Items[j].DetailsURL,
			})
		}
	}

	// Build Trends
	for i := range trendsHistory.Builds.BuildItems {
		for j := range trendsHistory.Builds.BuildItems[i].ItemsList {
			metricsData = append(metricsData, MetricData{
				"entity_name":                                   "SauceLabs",
				"event_type":                                    "SauceLabs",
				"provider":                                      "saucelabs",
				"saucelabs.name":                                trendsHistory.Builds.BuildItems[i].OwnerName,
				"saucelabs.trendsHistory.buildName":             trendsHistory.Builds.BuildItems[i].BuildName,
				"saucelabs.trendsHistory.buildTestsCount":       trendsHistory.Builds.BuildItems[i].TestsCount,
				"saucelabs.trendsHistory.buildDuration":         trendsHistory.Builds.BuildItems[i].Duration,
				"saucelabs.trendsHistory.buildDurationAbsolute": trendsHistory.Builds.BuildItems[i].DurationAbsolute,
				"saucelabs.trendsHistory.buildDurationTestMax":  trendsHistory.Builds.BuildItems[i].DurationTestMax,
				"saucelabs.trendsHistory.buildStartTime":        trendsHistory.Builds.BuildItems[i].StartTime,
				"saucelabs.trendsHistory.buildEndTime":          trendsHistory.Builds.BuildItems[i].EndTime,
				"saucelabs.trendsHistory.testID":                trendsHistory.Builds.BuildItems[i].ItemsList[j].ID,
				"saucelabs.trendsHistory.testOwner":             trendsHistory.Builds.BuildItems[i].ItemsList[j].Owner,
				"saucelabs.trendsHistory.testName":              trendsHistory.Builds.BuildItems[i].ItemsList[j].Name,
				"saucelabs.trendsHistory.testBuild":             trendsHistory.Builds.BuildItems[i].ItemsList[j].Build,
				"saucelabs.trendsHistory.testCreationTime":      trendsHistory.Builds.BuildItems[i].ItemsList[j].CreationTime,
				"saucelabs.trendsHistory.testStartTime":         trendsHistory.Builds.BuildItems[i].ItemsList[j].StartTime,
				"saucelabs.trendsHistory.testEndTime":           trendsHistory.Builds.BuildItems[i].ItemsList[j].EndTime,
				"saucelabs.trendsHistory.testDuration":          trendsHistory.Builds.BuildItems[i].ItemsList[j].Duration,
				"saucelabs.trendsHistory.testStatus":            trendsHistory.Builds.BuildItems[i].ItemsList[j].Status,
				"saucelabs.trendsHistory.testError":             trendsHistory.Builds.BuildItems[i].ItemsList[j].Error,
				"saucelabs.trendsHistory.testOS":                trendsHistory.Builds.BuildItems[i].ItemsList[j].OS,
				"saucelabs.trendsHistory.testBrowser":           trendsHistory.Builds.BuildItems[i].ItemsList[j].Browser,
				"saucelabs.trendsHistory.testDetailsURL":        trendsHistory.Builds.BuildItems[i].ItemsList[j].DetailsURL,
			})
		}
	}
	return metricsData, nil
}

func getHistoryDate(userHistory History, index int) time.Time {
	var year int
	var month int
	var day int
	var err error
	r, _ := regexp.Compile("([0-9]{4})+[-]+([0-9]{1,2})+[-]+([0-9]{1,2})")
	if r.MatchString(userHistory.Usage[index][0].(string)) {

		year, err = strconv.Atoi(userHistory.Usage[index][0].(string)[:4])
		if err != nil {
			fmt.Println("Year Convert Error")
		}

		check := userHistory.Usage[index][0].(string)[6:7]
		if check == "-" {
			month, err = strconv.Atoi(userHistory.Usage[index][0].(string)[5:6])
			if err != nil {
				fmt.Println("Month Convert Error")
			}

			day, err = strconv.Atoi(userHistory.Usage[index][0].(string)[7:])
			if err != nil {
				fmt.Println("Month Convert Error")
			}
		} else {
			month, err = strconv.Atoi(userHistory.Usage[index][0].(string)[5:7])
			if err != nil {
				fmt.Println("Month Convert Error")
			}

			day, err = strconv.Atoi(userHistory.Usage[index][0].(string)[8:])
			if err != nil {
				fmt.Println("Month Convert Error")
			}
		}
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	}
	log.Fatal("Error parsing users history date")
	return time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)
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

func getPathURL(startDateString string, endDateString string, path string) Path {
	pathURL := Path{
		Path: path,
		Parameter: []Parameter{
			Parameter{
				key:   "start",
				value: startDateString + "Z",
			},
			Parameter{
				key:   "end",
				value: endDateString + "Z",
			},
			Parameter{
				key:   "scope",
				value: "organization",
			},
		},
	}
	return pathURL
}

func validateConfig(config SauceConfig) error {
	if config.SauceAPIUser == "" && config.SauceAPIKey == "" {
		return fmt.Errorf("Config Yaml is missing SAUCE_API_USER and SAUCE_API_KEY values. Please check the config to continue")
	}
	if config.SauceAPIUser == "" {
		return fmt.Errorf("Config Yaml is missing SAUCE_API_USER value. Please check the config to continue")
	}
	if config.SauceAPIKey == "" {
		return fmt.Errorf("Config Yaml is missing SAUCE_API_KEY value. Please check the config to continue")
	}
	return nil
}
