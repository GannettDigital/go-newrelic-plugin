package collectors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/GannettDigital/paas-api-utils/utilsHTTP"
	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"
)

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

func executeAndDecode(runner utilsHTTP.HTTPRunner, httpReq http.Request, record interface{}) error {
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
	return json.NewDecoder(bytes.NewBuffer(data)).Decode(&record)
}

func listNodes(config RabbitmqConfig, runner utilsHTTP.HTTPRunner) (nodeRecords []NodeInfo, err error) {
	rabbitmqNodeStatsURI := fmt.Sprintf("%v:%v/%v", config.RabbitMQHost, config.RabbitMQPort, "api/nodes")
	httpReq, err := http.NewRequest("GET", rabbitmqNodeStatsURI, bytes.NewBuffer([]byte("")))
	if err != nil {
		log.WithFields(logrus.Fields{
			"rabbitmqNodeStatsURI": rabbitmqNodeStatsURI,
			"error":                err,
		}).Error("Encountered error creating http.NewRequest")
		return []NodeInfo{}, err
	}
	httpReq.SetBasicAuth(config.RabbitMQUser, config.RabbitMQPassword)
	err = executeAndDecode(runner, *httpReq, &nodeRecords)
	if err != nil {
		return []NodeInfo{}, err
	}

	return nodeRecords, nil
}

func listQueues(config RabbitmqConfig, runner utilsHTTP.HTTPRunner) (queueRecords []QueueInfo, err error) {
	rabbitmqQueuesStatsURI := fmt.Sprintf("%v:%v/%v", config.RabbitMQHost, config.RabbitMQPort, "api/queues")
	httpReq, err := http.NewRequest("GET", rabbitmqQueuesStatsURI, bytes.NewBuffer([]byte("")))
	if err != nil {
		log.WithFields(logrus.Fields{
			"rabbitmqQueuesStatsURI": rabbitmqQueuesStatsURI,
			"error":                  err,
		}).Error("Encountered error creating http.NewRequest")
		return []QueueInfo{}, err
	}
	httpReq.SetBasicAuth(config.RabbitMQUser, config.RabbitMQPassword)
	err = executeAndDecode(runner, *httpReq, &queueRecords)
	if err != nil {
		return []QueueInfo{}, err
	}
	return queueRecords, nil
}

func getRabbitmqStatus(config RabbitmqConfig, runner utilsHTTP.HTTPRunner) ([]map[string]interface{}, error) {
	NodesResponse, err := listNodes(config, runner)
	if err != nil {
		log.WithFields(logrus.Fields{
			"rabbitConfig": config,
			"error":        err,
		}).Error("Encountered error querying Nodes")
		return make([]map[string]interface{}, 0), err
	}
	Stats := make([]map[string]interface{}, 0)
	for _, Node := range NodesResponse {
		Stats = append(Stats, map[string]interface{}{
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

	QueuesResponse, err := listQueues(config, runner)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Encountered error querying Queues")
		return make([]map[string]interface{}, 0), err
	}
	for _, Queue := range QueuesResponse {
		Stats = append(Stats, map[string]interface{}{
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

//RabbitmqCollector gets the rabbits stats.
func RabbitmqCollector(config Config, stats chan<- []map[string]interface{}) {
	var runner utilsHTTP.HTTPRunnerImpl
	var rabbitConf RabbitmqConfig
	err := mapstructure.Decode(config.Collectors["rabbitmq"].CollectorConfig, &rabbitConf)
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("Unable to decode nginx config into NginxConfig object")

		close(stats)
	}
	rabbitResponses, getStatsError := getRabbitmqStatus(rabbitConf, runner)
	if getStatsError != nil {
		close(stats)
		return
	}
	stats <- rabbitResponses
}
