package rabbitmq

import (
	"reflect"
	"testing"

	fake "github.com/GannettDigital/paas-api-utils/utilsHTTP/fake"
	"github.com/franela/goblin"
	"github.com/sirupsen/logrus"
)

var rabbitMqFakeConfig RabbitmqConfig

func init() {
	rabbitMqFakeConfig = RabbitmqConfig{
		rabbitmqHost:     "http://localhost",
		rabbitmqPassword: "secure",
		rabbitmqPort:     "15672",
		rabbitmqUser:     "admin",
	}
}

func TestListQueues(t *testing.T) {
	g := goblin.Goblin(t)
	resultSlice := make([]QueueInfo, 1)
	resultSlice[0] = QueueInfo{
		Name:                   "TheTestQueue",
		Vhost:                  "TheTestVhost",
		Durable:                false,
		Node:                   "rabbit@rabbit-1",
		Memory:                 55240,
		Consumers:              1,
		Policy:                 "ha-all",
		MessagesBytes:          0,
		Messages:               0,
		MessagesReady:          0,
		MessagesUnacknowledged: 0,
	}

	var tests = []struct {
		HTTPRunner      fake.HTTPResult
		TestDescription string
		ExpectedResult  []QueueInfo
	}{
		{
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					fake.Result{
						Method: "GET",
						URI:    "/api/queues",
						Code:   200,
						Data:   []byte("[  { \"messages\": 0,\"messages_ready\": 0,\"messages_unacknowledged\": 0,\"policy\": \"ha-all\",\"consumers\": 1,\"memory\": 55240,\"message_bytes\": 0,\"name\": \"TheTestQueue\",\"vhost\": \"TheTestVhost\",\"durable\": false,\"node\": \"rabbit@rabbit-1\"  }]"),
					},
				},
			},
			TestDescription: "Successfully GET Queue List",
			ExpectedResult:  resultSlice,
		},
	}

	for _, test := range tests {
		g.Describe("listQueues()", func() {
			g.It(test.TestDescription, func() {
				runner = test.HTTPRunner
				result, err := listQueues(logrus.New(), rabbitMqFakeConfig)
				g.Assert(reflect.DeepEqual(err, nil)).Equal(true)
				g.Assert(reflect.DeepEqual(result, resultSlice)).Equal(true)
			})
		})
	}
}

func TestListNodes(t *testing.T) {
	g := goblin.Goblin(t)
	resultSlice := make([]NodeInfo, 2)
	resultSlice[0] = NodeInfo{
		Name:           "rabbit@rabbit-1",
		FdUsed:         29,
		FdTotal:        65536,
		SocketsUsed:    1,
		SocketsTotal:   58890,
		MemUsed:        61525440,
		DiskFree:       1932713984,
		RunQueueLength: 0,
		Processors:     1,
		Uptime:         22237082,
	}
	resultSlice[1] = NodeInfo{
		Name:           "rabbit@rabbit-2",
		FdUsed:         29,
		FdTotal:        65536,
		SocketsUsed:    1,
		SocketsTotal:   58890,
		MemUsed:        61525440,
		DiskFree:       1932713984,
		RunQueueLength: 0,
		Processors:     1,
		Uptime:         22237082,
	}

	var tests = []struct {
		HTTPRunner      fake.HTTPResult
		TestDescription string
		ExpectedResult  []NodeInfo
	}{
		{
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					fake.Result{
						Method: "GET",
						URI:    "/api/nodes",
						Code:   200,
						Data:   []byte("[{\"disk_free\":1932713984,\"fd_used\":29,\"mem_used\":61525440,\"sockets_used\":1,\"fd_total\":65536,\"sockets_total\":58890,\"proc_total\":32768,\"rates_mode\":\"basic\",\"uptime\":22237082,\"run_queue\":0,\"processors\":1,\"name\":\"rabbit@rabbit-1\"},{\"disk_free\":1932713984,\"fd_used\":29,\"mem_used\":61525440,\"sockets_used\":1,\"fd_total\":65536,\"sockets_total\":58890,\"proc_total\":32768,\"rates_mode\":\"basic\",\"uptime\":22237082,\"run_queue\":0,\"processors\":1,\"name\":\"rabbit@rabbit-2\"}]"),
					},
				},
			},
			TestDescription: "Successfully GET Node List",
			ExpectedResult:  resultSlice,
		},
	}

	for _, test := range tests {
		g.Describe("listNodes()", func() {
			g.It(test.TestDescription, func() {
				runner = test.HTTPRunner
				result, err := listNodes(logrus.New(), rabbitMqFakeConfig)
				g.Assert(reflect.DeepEqual(err, nil)).Equal(true)
				g.Assert(reflect.DeepEqual(result, resultSlice)).Equal(true)
			})
		})
	}
}

func TestGetRabbitmqStatus(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		HTTPRunner      fake.HTTPResult
		TestDescription string
	}{
		{
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					fake.Result{
						Method: "GET",
						URI:    "/api/nodes",
						Code:   200,
						Data:   []byte("[{\"disk_free\":1932713984,\"fd_used\":29,\"mem_used\":61525440,\"sockets_used\":1,\"fd_total\":65536,\"sockets_total\":58890,\"proc_total\":32768,\"rates_mode\":\"basic\",\"uptime\":22237082,\"run_queue\":0,\"processors\":1,\"name\":\"rabbit@rabbit-1\"},{\"disk_free\":1932713984,\"fd_used\":29,\"mem_used\":61525440,\"sockets_used\":1,\"fd_total\":65536,\"sockets_total\":58890,\"proc_total\":32768,\"rates_mode\":\"basic\",\"uptime\":22237082,\"run_queue\":0,\"processors\":1,\"name\":\"rabbit@rabbit-2\"}]"),
					},
					fake.Result{
						Method: "GET",
						URI:    "/api/queues",
						Code:   200,
						Data:   []byte("[  { \"messages\": 0,\"messages_ready\": 0,\"messages_unacknowledged\": 0,\"policy\": \"ha-all\",\"consumers\": 1,\"memory\": 55240,\"message_bytes\": 0,\"name\": \"TheTestQueue\",\"vhost\": \"TheTestVhost\",\"durable\": false,\"node\": \"rabbit@rabbit-1\"  }]"),
					},
				},
			},
			TestDescription: "Get RabbitMqStatus",
		},
	}

	for _, test := range tests {
		g.Describe("listNodes()", func() {
			g.It(test.TestDescription, func() {
				runner = test.HTTPRunner
				result, err := getRabbitmqStatus(logrus.New(), rabbitMqFakeConfig)
				g.Assert(reflect.DeepEqual(err, nil)).Equal(true)
				g.Assert(len(result) == 3).Equal(true)
			})
		})
	}
}
