package mysql

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/GannettDigital/go-newrelic-plugin/helpers"
	"github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const NAME string = "mysql"
const PROVIDER string = "mysql"
const PROTOCOL_VERSION string = "1"
const PLUGIN_VERSION string = "1.0.0"
const STATUS string = "OK"

//mysqlConfig is the keeper of the config
type mysqlConfig struct {
	host     string
	port     string
	user     string
	password string
	database string
	queries  string
	prefixes string
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

var log *logrus.Logger

var config = mysqlConfig{
	host:     os.Getenv("HOST"),
	port:     os.Getenv("PORT"),
	user:     os.Getenv("USER"),
	password: os.Getenv("PASSWORD"),
	database: os.Getenv("DATABASE"),
	queries:  os.Getenv("QUERIES"),
	prefixes: os.Getenv("PREFIXES"),
}

func Run(logger *logrus.Logger, prettyPrint bool, version string) {
	log = logger
	// Initialize the output structure
	var data = PluginData{
		Name:            NAME,
		PluginVersion:   PLUGIN_VERSION,
		ProtocolVersion: PROTOCOL_VERSION,
		Status:          STATUS,
		Metrics:         make([]MetricData, 0),
		Inventory:       make(map[string]InventoryData),
		Events:          make([]EventData, 0),
	}

	validateConfig()

	db, err := sql.Open("mysql", generateDSN())
	if err != nil {
		log.WithError(err).Error(fmt.Sprintf("getMetric: Cannot connect to mysql %s:%s", config.host, config.port))
		return 
	}
	defer db.Close()

	metric, err := getMetrics(db)
	if err != nil {
		data.Status = err.Error()
	}
	data.Metrics = append(data.Metrics, metric)
	fatalIfErr(helpers.OutputJSON(data, prettyPrint), "OutputJSON error")
}

func getMetrics(db *sql.DB) (map[string]interface{}, error) {

	metrics := map[string]interface{}{
		"event_type": "DatastoreSample",
		"provider":   PROVIDER,
	}

	for _, query := range strings.Split(config.queries, ";") {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}
		rows, err := db.Query(query)
		if err != nil {
			log.WithError(err).Warn(" query; " + query)
			continue
		}
		defer rows.Close()

		cols, _ := rows.Columns()
		rawResult := make([][]byte, len(cols))
		result := make([]string, len(cols))
		dest := make([]interface{}, len(cols)) // A temporary interface{} slice
		for i, _ := range rawResult {
			dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
		}

		for rows.Next() {
			err = rows.Scan(dest...)
			if err != nil {
				log.WithError(err).Warn(fmt.Sprintf("Failed to scan row. Query: %s", query))
				continue
			}

			if len(rawResult) != 2 {
				for i, raw := range rawResult {
					if raw == nil {
						result[i] = "\\N"
					} else {
						result[i] = string(raw)
					}
				}
				log.Warn(fmt.Sprintf("Unknown query result: query %s result: %#v\n", query, result))
			} else {
				name := metricName(string(rawResult[0]))
				metrics[name] = asValue(string(rawResult[1]))
			}
		}
	}
	return metrics, nil
}

func metricName(metric string) string {
	log.Debug(fmt.Sprintf("metricName: metric: %s", metric))
	result := fmt.Sprintf("mysql.%s", camelCase(metric))
	log.Debug(fmt.Sprintf("metricName: result3: %s", result))
	return result
}

var camelingRegex = regexp.MustCompile("[0-9A-Za-z.]+")

func camelCase(src string) string {
	log.Debug(fmt.Sprintf("camelCase: src: %s", src))
	src = strings.Replace(src, ":", ".", -1)
	src = fixPrefix(src)
	byteSrc := []byte(src)
	chunks := camelingRegex.FindAll(byteSrc, -1)
	for idx, val := range chunks {
		if idx > 0 {
			chunks[idx] = bytes.Title(val)
		}
	}
	result := string(bytes.Join(chunks, nil))
	log.Debug(fmt.Sprintf("camelCase: result: %s", result))
	return result
}

func fixPrefix(src string) string {
	for _, prefix := range strings.Split(config.prefixes, " ") {
		if strings.HasPrefix(src, prefix) {
			src = strings.Replace(src, "_", ".", 1)
			return src
		}
	}
	return src
}

func asValue(value string) interface{} {
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}

	if b, err := strconv.ParseBool(value); err == nil {
		return b
	}
	return value
}

func generateDSN() string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", config.user, config.password, config.host, config.port, config.database)
	log.Debug("generateDSN: %s", dsn)
	return dsn
}

func validateConfig() {
	if config.host == "" {
		log.Fatal("Config Yaml is missing HOST value. Please check the config to continue")
	}
	if config.port == "" {
		log.Fatal("Config Yaml is missing PORT value. Please check the config to continue")
	}
	if config.user == "" {
		log.Fatal("Config Yaml is missing USER value. Please check the config to continue")
	}
	if config.password == "" {
		log.Fatal("Config Yaml is missing PASSWORD value. Please check the config to continue")
	}
	if config.database == "" {
		log.Fatal("Config Yaml is missing DATABASE value. Please check the config to continue")
	}
	if config.queries == "" {
		log.Fatal("Config Yaml is missing QUERIES value. Please check the config to continue")
	}
	if config.prefixes == "" {
		log.Fatal("Config Yaml is missing PREFIXES value. Please check the config to continue")
	}
}

func fatalIfErr(err error, msg string) {
	if err != nil {
		log.WithError(err).Fatal(msg)
	}
}
