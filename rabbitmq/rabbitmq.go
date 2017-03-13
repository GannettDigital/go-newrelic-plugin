package rabbitmq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/GannettDigital/go-newrelic-plugin/types"
	"github.com/GannettDigital/paas-api-utils/utilsHTTP"
	"github.com/Sirupsen/logrus"
)

var runner utilsHTTP.HTTPRunner

const NAME string = "rabbitmq"
const PROVIDER string = "rabbitmq" //we might want to make this an env tied to nginx version or app name maybe...
const PROTOCOL_VERSION string = "1"
const EVENT_TYPE string = "QueueSample"

//RabbitmqConfig is the keeper of the config
type RabbitmqConfig struct {
	rabbitmqUser     string
	rabbitmqPassword string
	rabbitmqPort     string
	rabbitmqHost     string
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

type NodeInfo struct {
	Name           string `json:"name"`
	FdUsed         int    `json:"fd_used"`
	FdTotal        int    `json:"fd_total"`
	SocketsUsed    int    `json:"sockets_used"`
	SocketsTotal   int    `json:"sockets_total"`
	MemUsed        int    `json:"mem_used"`
	DiskFree       int    `json:"disk_free"`
	DiskFreeLimit  int    `json:"disk_free_limit"`
	RunQueueLength uint32 `json:"run_queue"`
	Processors     uint32 `json:"processors"`
	Uptime         uint64 `json:"uptime"`
}

type QueueInfo struct {
	// Queue name
	Name string `json:"name"`
	// Virtual host this queue belongs to
	Vhost string `json:"vhost"`
	// Is this queue durable?
	Durable bool `json:"durable"`
	// RabbitMQ node that hosts master for this queue
	Node string `json:"node"`
	// Total amount of RAM used by this queue
	Memory int64 `json:"memory"`
	// How many consumers this queue has
	Consumers int `json:"consumers"`
	// Policy applied to this queue, if any
	Policy string `json:"policy"`
	// Total bytes of messages in this queues
	MessagesBytes int64 `json:"message_bytes"`
	// Total number of messages in this queue
	Messages int `json:"messages"`
	// Number of messages ready to be delivered
	MessagesReady int `json:"messages_ready"`
	// Number of messages delivered and pending acknowledgements from consumers
	MessagesUnacknowledged int `json:"messages_unacknowledged"`
}

func init() {
	runner = utilsHTTP.HTTPRunnerImpl{}
}

func executeAndDecode(log *logrus.Logger, httpReq http.Request, record interface{}) error {
	code, data, err := runner.CallAPI(log, nil, &httpReq, &http.Client{})
	if err != nil || code != 200 {
		log.WithFields(logrus.Fields{
			"code":    code,
			"data":    string(data),
			"httpReq": httpReq,
			"error":   err,
		}).Error("Encountered error calling CallAPI")
		return err
	}
	return json.Unmarshal(data, &record)
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

func validateConfig(log *logrus.Logger, config RabbitmqConfig) {
	if config.rabbitmqHost == "" || config.rabbitmqPassword == "" || config.rabbitmqPort == "" || config.rabbitmqUser == "" {
		log.Fatal("Config Yaml is missing values. Please check the config to continue")
	}
}

func fatalIfErr(log *logrus.Logger, err error) {
	if err != nil {
		log.WithError(err).Fatal("can't continue")
	}
}

func Run(log *logrus.Logger, opts types.Opts, version string) {

	// Initialize the output structure
	var data = PluginData{
		Name:            NAME,
		ProtocolVersion: PROTOCOL_VERSION,
		PluginVersion:   version,
		Inventory:       make(map[string]InventoryData),
		Metrics:         make([]MetricData, 0),
		Events:          make([]EventData, 0),
	}

	var config = RabbitmqConfig{
		rabbitmqUser:     os.Getenv("RABBITMQ_USER"),
		rabbitmqPassword: os.Getenv("RABBITMQ_PASSWORD"),
		rabbitmqPort:     os.Getenv("RABBITMQ_PORT"),
		rabbitmqHost:     os.Getenv("RABBITMQ_HOST"),
	}
	validateConfig(log, config)

	metrics, err := getRabbitmqStatus(log, config)
	fatalIfErr(log, err)

	data.Metrics = append(data.Metrics, metrics...)
	fatalIfErr(log, OutputJSON(data, opts.PrettyPrint))
}

func listNodes(log *logrus.Logger, config RabbitmqConfig) (nodeRecords []NodeInfo, err error) {
	rabbitmqNodeStatsURI := fmt.Sprintf("%v:%v/%v", config.rabbitmqHost, config.rabbitmqPort, "api/nodes")
	httpReq, err := http.NewRequest("GET", rabbitmqNodeStatsURI, bytes.NewBuffer([]byte("")))
	if err != nil {
		log.WithFields(logrus.Fields{
			"rabbitmqNodeStatsURI": rabbitmqNodeStatsURI,
			"error":                err,
		}).Error("Encountered error creating http.NewRequest")
		return []NodeInfo{}, err
	}
	httpReq.SetBasicAuth(config.rabbitmqUser, config.rabbitmqPassword)
	err = executeAndDecode(log, *httpReq, &nodeRecords)
	if err != nil {
		return []NodeInfo{}, err
	}

	return nodeRecords, nil
}

func listQueues(log *logrus.Logger, config RabbitmqConfig) (queueRecords []QueueInfo, err error) {
	rabbitmqQueuesStatsURI := fmt.Sprintf("%v:%v/%v", config.rabbitmqHost, config.rabbitmqPort, "api/queues")
	httpReq, err := http.NewRequest("GET", rabbitmqQueuesStatsURI, bytes.NewBuffer([]byte("")))
	if err != nil {
		log.WithFields(logrus.Fields{
			"rabbitmqQueuesStatsURI": rabbitmqQueuesStatsURI,
			"error":                  err,
		}).Error("Encountered error creating http.NewRequest")
		return []QueueInfo{}, err
	}
	httpReq.SetBasicAuth(config.rabbitmqUser, config.rabbitmqPassword)
	err = executeAndDecode(log, *httpReq, &queueRecords)
	if err != nil {
		return []QueueInfo{}, err
	}
	return queueRecords, nil
}

func getRabbitmqStatus(log *logrus.Logger, config RabbitmqConfig) ([]MetricData, error) {
	NodesResponse, err := listNodes(log, config)
	if err != nil {
		log.WithFields(logrus.Fields{
			"rabbitConfig": config,
			"error":        err,
		}).Error("Encountered error querying Nodes")
		return make([]MetricData, 0), err
	}
	Stats := make([]MetricData, 0)
	for _, Node := range NodesResponse {
		Stats = append(Stats, MetricData{
			"event_type":                  EVENT_TYPE,
			"provider":                    PROVIDER,
			"rabbitmq.node.name":          Node.Name,
			"rabbitmq.node.fd_used":       Node.FdUsed,
			"rabbitmq.node.fd_total":      Node.FdTotal,
			"rabbitmq.node.mem_used":      Node.MemUsed,
			"rabbitmq.node.sockets_used":  Node.SocketsUsed,
			"rabbitmq.node.sockets_total": Node.SocketsTotal,
			"rabbitmq.node.run_queue":     Node.RunQueueLength,
			"rabbitmq.node.processors":    Node.Processors,
		})
	}

	QueuesResponse, err := listQueues(log, config)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Encountered error querying Queues")
		return make([]MetricData, 0), err
	}
	for _, Queue := range QueuesResponse {
		Stats = append(Stats, MetricData{
			"event_type":                             EVENT_TYPE,
			"provider":                               PROVIDER,
			"rabbitmq.queue.name":                    Queue.Name,
			"rabbitmq.queue.vhost":                   Queue.Vhost,
			"rabbitmq.queue.durable":                 Queue.Durable,
			"rabbitmq.queue.memory":                  Queue.Memory,
			"rabbitmq.queue.consumers":               Queue.Consumers,
			"rabbitmq.queue.messages_bytes":          Queue.MessagesBytes,
			"rabbitmq.queue.messages":                Queue.Messages,
			"rabbitmq.queue.messages_ready":          Queue.MessagesReady,
			"rabbitmq.queue.messages_unacknowledged": Queue.MessagesUnacknowledged,
		})
	}

	//return Stats, nil
	return Stats, nil
}
