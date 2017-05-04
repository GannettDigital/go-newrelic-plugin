package memcached

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/franela/goblin"
)

var fakeConfig MemcachedConfig

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	log = logrus.New()
	var l = GetListener()
	var port = l.Addr().(*net.TCPAddr).Port
	// Spin the server out in a go routine
	go Server(l)

	fakeConfig = MemcachedConfig{
		MemcachedHost: "localhost",
		MemcachedPort: strconv.Itoa(port),
		Commands:      "stats , stats settings , stats items , stats sizes , stats slabs , stats conns",
	}

}
func GetListener() *net.TCPListener {
	// Using port 0 tells the OS to give us an open port, this should always 'just work'
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	return l
}

func Server(l *net.TCPListener) {
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

// Handles 'incoming' requests.
func handleRequest(conn net.Conn) {
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	conn.Write([]byte("STAT slab1:chunks_per_page 13"))
	conn.Close()
}

func TestGetMetric(t *testing.T) {
	g := goblin.Goblin(t)
	var tests = []struct {
		TestDescription string
		result          map[string]interface{}
	}{
		{
			TestDescription: "Should get metrics without error",
			result:          map[string]interface{}{"event_type": "DatastoreSample", "provider": "memcached", "memcached.slab1.chunksPerPage": 13},
		},
	}
	for _, test := range tests {
		g.Describe("getMetric)", func() {
			g.It(test.TestDescription, func() {
				metric, _ := getMetric(fakeConfig)
				log.Debug(metric)
				g.Assert(metric).Equal(test.result)
			})
		})
	}
}

func TestMetricName(t *testing.T) {
	g := goblin.Goblin(t)
	var tests = []struct {
		TestDescription string
		command         string
		metric          string
		result          string
	}{
		{
			TestDescription: "Should get metric name without error",
			command:         "stats slabs",
			metric:          "slab1:chunks_per_page",
			result:          "memcached.slabs.slab1.chunksPerPage",
		},
		{
			TestDescription: "Should get metric name without error",
			command:         "stats",
			metric:          "time_in_listen_disabled_us",
			result:          "memcached.timeInListenDisabledUs",
		},
	}
	for _, test := range tests {
		g.Describe("getMetricName)", func() {
			g.It(test.TestDescription, func() {
				name := metricName(test.command, test.metric)
				g.Assert(name).Equal(test.result)
			})
		})
	}
}

func TestCamelCase(t *testing.T) {
	g := goblin.Goblin(t)
	var tests = []struct {
		TestDescription string
		src             string
		result          string
	}{
		{
			TestDescription: "Should convert to camel case without error",
			src:             "one_two_three_four",
			result:          "oneTwoThreeFour",
		},
		{
			TestDescription: "Should convert to camel case with colon without error",
			src:             "one:two_three_four",
			result:          "one.twoThreeFour",
		},
	}
	for _, test := range tests {
		g.Describe("camelCase)", func() {
			g.It(test.TestDescription, func() {
				camel := camelCase(test.src)
				g.Assert(camel).Equal(test.result)
			})
		})
	}
}

func TestAsValue(t *testing.T) {
	g := goblin.Goblin(t)
	var tests = []struct {
		TestDescription string
		value           string
		result          interface{}
	}{
		{
			TestDescription: "Should convert string to string without error",
			value:           "string",
			result:          "string",
		},
		{
			TestDescription: "Should convert string to int without error",
			value:           "1",
			result:          1,
		},
		{
			TestDescription: "Should convert string to float without error",
			value:           "1.0",
			result:          1.0,
		},
		{
			TestDescription: "Should convert string to boolean without error",
			value:           "true",
			result:          true,
		},
	}
	for _, test := range tests {
		g.Describe("asValue)", func() {
			g.It(test.TestDescription, func() {
				value := asValue(test.value)
				g.Assert(value).Equal(test.result)
			})
		})
	}
}
