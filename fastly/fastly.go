package fastly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/GannettDigital/go-newrelic-plugin/helpers"
	"github.com/GannettDigital/paas-api-utils/utilsHTTP"
	"github.com/Sirupsen/logrus"
)

var runner utilsHTTP.HTTPRunner

// NAME - name of plugin
const NAME string = "fastly"

// PROVIDER -
const PROVIDER string = "fastly" //we might want to make this an env tied to nginx version or app name maybe...

// ProtocolVersion -
const ProtocolVersion string = "1"

// Fastly Stats endpoint
const FastlyStatsEndpoint = "https://rt.fastly.com/v1/"

// FastlyConfig is the keeper of the config
type Config struct {
	FastlyAPIKey string
	ServiceID    string
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

type FastlyRealTimeDataV1 struct {
	Data []FastlyDataObjects `json:"Data"`
}

type FastlyDataObjects struct {
	Datacenter map[string]FastlyStats `json:"datacenter"`
	Aggregated FastlyStats            `json:"aggregated"`
}

type FastlyStats struct {
	Requests         int     `json:"requests"`
	HeaderSize       int     `json:"header_size"`
	BodySize         int     `json:"body_size"`
	ReqHeaderBytes   int     `json:"req_header_bytes"`
	RespHeaderBytes  int     `json:"resp_header_bytes"`
	RespBodyBytes    int     `json:"resp_body_bytes"`
	BeReqHeaderBytes int     `json:"bereq_header_bytes"`
	Tls              int     `json:"tls"`
	Shield           int     `json:"shield"`
	Http2            int     `json:"http2"`
	Status2xx        int     `json:"status_2xx"`
	Status3xx        int     `json:"status_3xx"`
	Status4xx        int     `json:"status_4xx"`
	Status5xx        int     `json:"status_5xx"`
	Status200        int     `json:"status_200"`
	Status301        int     `json:"status_301"`
	Status302        int     `json:"status_302"`
	Status304        int     `json:"status_304"`
	Hits             int     `json:"hits"`
	Miss             int     `json:"miss"`
	Pass             int     `json:"pass"`
	Synth            int     `json:"synth"`
	Errors           int     `json:"errors"`
	HitsTime         float64 `json:"hits_time"`
	MissTime         float64 `json:"miss_time"`
}

func init() {
	runner = utilsHTTP.HTTPRunnerImpl{}
}

func Run(log *logrus.Logger, prettyPrint bool, version string) {

	// Initialize the output structure
	var data = PluginData{
		Name:            NAME,
		ProtocolVersion: ProtocolVersion,
		PluginVersion:   version,
		Inventory:       make(map[string]InventoryData),
		Metrics:         make([]MetricData, 0),
		Events:          make([]EventData, 0),
	}

	var fastlyConf = Config{
		FastlyAPIKey: os.Getenv("FASTLY_API_KEY"),
		ServiceID:    os.Getenv("SERVICE_ID"),
	}
	validateConfig(log, fastlyConf)

	fastlyStats := getFastlyStats(log, fastlyConf)

	// // loop over datacenter items
	for _, dataItem := range fastlyStats.Data {
		for datacenter, datacenterStats := range dataItem.Datacenter {
			data.Metrics = append(data.Metrics, convertToNrMetric(datacenterStats, datacenter, fastlyConf, log))
		}
		// push the aggregated type onto the stack
		data.Metrics = append(data.Metrics, convertToNrMetric(dataItem.Aggregated, "aggregated", fastlyConf, log))
	}

	fatalIfErr(log, helpers.OutputJSON(data, prettyPrint))
}

func convertToNrMetric(stats FastlyStats, dataCenter string, config Config, log *logrus.Logger) map[string]interface{} {

	return map[string]interface{}{
		"event_type":              "LoadBalancerSample",
		"provider":                PROVIDER,
		"fastly.serviceId":        config.ServiceID,
		"fastly.datacenter":       dataCenter,
		"fastly.requests":         stats.Requests,
		"fastly.headerSize":       stats.HeaderSize,
		"fastly.bodySize":         stats.BodySize,
		"fastly.reqHeaderBytes":   stats.ReqHeaderBytes,
		"fastly.respHeaderBytes":  stats.RespHeaderBytes,
		"fastly.bereqHeaderBytes": stats.BeReqHeaderBytes,
		"fastly.tls":              stats.Tls,
		"fastly.shield":           stats.Shield,
		"fastly.http2":            stats.Http2,
		"fastly.status.2xx":       stats.Status2xx,
		"fastly.status.3xx":       stats.Status3xx,
		"fastly.status.4xx":       stats.Status4xx,
		"fastly.status.5xx":       stats.Status5xx,
		"fastly.status.200":       stats.Status200,
		"fastly.status.301":       stats.Status301,
		"fastly.status.302":       stats.Status302,
		"fastly.status.304":       stats.Status304,
		"fastly.hits":             stats.Hits,
		"fastly.pass":             stats.Pass,
		"fastly.synth":            stats.Synth,
		"fastly.errors":           stats.Errors,
		"fastly.hitTime":          stats.HitsTime,
		"fastly.missTime":         stats.MissTime,
	}
}

func validateConfig(log *logrus.Logger, fastlyConf Config) {
	if fastlyConf.FastlyAPIKey == "" || fastlyConf.ServiceID == "" {
		log.Fatal("Config Yaml is missing values. Please check the config to continue")
	}
}

func fatalIfErr(log *logrus.Logger, err error) {
	if err != nil {
		log.WithError(err).Fatal("can't continue")
	}
}

func getFastlyStats(log *logrus.Logger, config Config) FastlyRealTimeDataV1 {
	fastlyStats := fmt.Sprintf("%vchannel/%v/ts/0", FastlyStatsEndpoint, config.ServiceID)
	httpReq, err := http.NewRequest("GET", fastlyStats, bytes.NewBuffer([]byte("")))
	httpReq.Header.Set("Fastly-Key", config.FastlyAPIKey)
	httpReq.Header.Set("Content-Type", "application/json")
	// http.NewRequest error
	fatalIfErr(log, err)
	code, data, err := runner.CallAPI(log, nil, httpReq, &http.Client{})
	if err != nil || code != 200 {
		fmt.Fprintln(os.Stderr, err.Error())
		log.WithFields(logrus.Fields{
			"code":             code,
			"data":             string(data),
			"httpReq":          httpReq,
			"FastlyEndpoint":   FastlyStatsEndpoint,
			"config.ServiceID": config.ServiceID,
			"error":            err,
		}).Fatal("Encountered error calling CallAPI")
		return FastlyRealTimeDataV1{}
	}

	var fastlyData FastlyRealTimeDataV1

	if err := json.Unmarshal(data, &fastlyData); err != nil {
		log.Panic("unable to unmarshal return data into fastly stats type: ", err.Error())
	}

	return fastlyData
}
