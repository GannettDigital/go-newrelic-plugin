package saucelabs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
)

// CollectorName - the name of this thing
const NAME string = "saucelabs"

// ProviderName - what app is sending the data
const PROVIDER string = "saucelabs"

// ProtocolVersion - nr-infra protocol version
const PROTOCOL_VERSION string = "1"

const url = "https://saucelabs.com/rest/v1/users/"

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

// UserMetric holds the user metrics
type User struct {
	UserName string `json:"username"`
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
	fmt.Println(metric)
	//data.Metrics = append(data.Metrics, metric)
	fatalIfErr(log, OutputJSON(data, prettyPrint))
}

func getMetric(log *logrus.Logger, config SauceConfig) string {

	client := http.Client{
		Timeout: time.Second * 2,
	}

	test := gerUserList(client, config)
	fmt.Println(test)
	return "test1"
}

func validateConfig(config SauceConfig) error {
	if config.SauceAPIUser != "" && config.SauceAPIKey == "" {
		return fmt.Errorf("You must also set SAUCE_API_KEY if SAUCE_API_USER is set")
	}
	if config.SauceAPIUser == "" && config.SauceAPIKey != "" {
		return fmt.Errorf("You must also set SAUCE_API_USER if SAUCE_API_KEY is set")
	}
	return nil
}

func fatalIfErr(log *logrus.Logger, err error) {
	if err != nil {
		log.WithError(err).Fatal("can't continue")
	}
}

func gerUserList(client http.Client, config SauceConfig) []User {
	userList := []User{}
	getUserListURL := url + config.SauceAPIUser + "/subaccounts"

	req, err := http.NewRequest(http.MethodGet, getUserListURL, nil)
	if err != nil {
		return nil
	}
	//set User-Agent to be a good internet citizen
	req.Header.Set("User-Agent", "GannettDigital-API")
	//set api key
	req.SetBasicAuth(config.SauceAPIUser, config.SauceAPIKey)

	res, errdo := client.Do(req)
	if errdo != nil {
		return nil
	}
	body, errread := ioutil.ReadAll(res.Body)
	if errread != nil {
		return nil
	}
	fmt.Println("test2")
	fmt.Println(body)

	//for loop to get user list
	// for _, name := range body {
	// 	userList = append(userList, body.username)
	// }

	// for i := 0; i < len(body); i++ {
	// 	userList.UserName = append(userList.UserName, body.username)
	// }

	return userList
}
