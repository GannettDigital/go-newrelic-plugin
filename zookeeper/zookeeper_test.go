package zookeeper

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/franela/goblin"
)

var fakeConfig Config

var conflistner *net.TCPListener
var mntrlistner *net.TCPListener

var confPort int
var mntrPort int

func init() {
	logrus.SetLevel(logrus.DebugLevel)

	// Spin the server out in a go routine for conf
	conflistner = GetListener()
	confPort = conflistner.Addr().(*net.TCPAddr).Port
	go serverConf(conflistner)

	// Spin the server out in a go routine for mntr
	mntrlistner = GetListener()
	mntrPort = mntrlistner.Addr().(*net.TCPAddr).Port
	go serverMntr(mntrlistner)

	fakeConfig = Config{
		ZK_HOST:       "localhost",
		ZK_CLIENTPORT: "0",
		ZK_TICKTIME:   "2000",
		ZK_DATADIR:    "/var/lib/zookeeper",
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

func serverConf(l *net.TCPListener) {
	for {
		// Listen for an incoming connection.
		conn, conn_err := l.Accept()
		if conn_err != nil {
			fmt.Println("Error accepting: ", conn_err.Error())
			os.Exit(1)
		}

		//Write fake FLW conf output
		buf := make([]byte, 1024)
		_, read_err := conn.Read(buf)
		if read_err != nil {
			fmt.Println("Error reading:", read_err.Error())
		}

		conn.Write([]byte("clientPort=" + strconv.Itoa(l.Addr().(*net.TCPAddr).Port) + "\ndataDir=" +
			fakeConfig.ZK_DATADIR + "\ntickTime=" + fakeConfig.ZK_TICKTIME +
			"\nmaxClientCnxns=60\nminSessionTimeout=4000\nmaxSessionTimeout=40000\nserverId=0"))
		conn.Close()
	}
}

func serverMntr(l *net.TCPListener) {
	for {
		// Listen for an incoming connection.
		conn, conn_err := l.Accept()
		if conn_err != nil {
			fmt.Println("Error accepting: ", conn_err.Error())
			os.Exit(1)
		}

		//Write fake FLW mntr output
		buf := make([]byte, 1024)
		_, read_err := conn.Read(buf)
		if read_err != nil {
			fmt.Println("Error reading:", read_err.Error())
		}
		conn.Write([]byte("zk_version\t3.4.10-39d3a4f269333c922ed3db283be479f9deacaa0f\nzk_avg_latency\t0\nzk_max_latency\t0\nzk_min_latency\t0\nzk_packets_received\t2\n" +
			"zk_packets_sent\t1\nzk_num_alive_connections\t1\nzk_outstanding_requests\t0\nzk_avg_latency\tstandalone\n" +
			"zk_znode_count\t19\nzk_watch_count\t0\nzk_ephemerals_count\t0\nzk_approximate_data_size\t270\n" +
			"zk_open_file_descriptor_count\t38\nzk_max_file_descriptor_count\t10240\n"))
		conn.Close()
	}
}

func TestFatalIfErrt(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputLog        *logrus.Logger
		InputErr        error
		TestDescription string
	}{
		{
			InputLog:        logrus.New(),
			InputErr:        nil,
			TestDescription: "Should successfully not exit on a nil error",
		},
	}

	for _, test := range tests {
		g.Describe("fatalIfErr()", func() {
			g.It(test.TestDescription, func() {
				fatalIfErr(test.InputLog, test.InputErr)
				g.Assert(true).IsTrue()
			})
		})
	}
}

func TestGetFLWmntr(t *testing.T) {
	fakeConfig.ZK_CLIENTPORT = strconv.Itoa(mntrPort)
	g := goblin.Goblin(t)
	var tests = []struct {
		ExpectedResult  string
		TestDescription string
	}{
		{
			ExpectedResult: "zk_version\t3.4.10-39d3a4f269333c922ed3db283be479f9deacaa0f\nzk_avg_latency\t0\nzk_max_latency\t0\nzk_min_latency\t0\nzk_packets_received\t2\n" +
				"zk_packets_sent\t1\nzk_num_alive_connections\t1\nzk_outstanding_requests\t0\nzk_avg_latency\tstandalone\n" +
				"zk_znode_count\t19\nzk_watch_count\t0\nzk_ephemerals_count\t0\nzk_approximate_data_size\t270\n" +
				"zk_open_file_descriptor_count\t38\nzk_max_file_descriptor_count\t10240\n",
			TestDescription: "Successfully got data",
		},
	}

	for _, test := range tests {
		g.Describe("getFLWmntr()", func() {
			g.It(test.TestDescription, func() {
				result := getFLWmntr(logrus.New(), fakeConfig)
				fmt.Println(result)
				g.Assert(reflect.DeepEqual(result, test.ExpectedResult)).Equal(true)
			})
		})
	}
}

func TestGetFLWconf(t *testing.T) {
	fakeConfig.ZK_CLIENTPORT = strconv.Itoa(confPort)
	g := goblin.Goblin(t)
	var tests = []struct {
		ExpectedResult  string
		TestDescription string
	}{
		{
			ExpectedResult: "clientPort=" + fakeConfig.ZK_CLIENTPORT + "\ndataDir=" +
				fakeConfig.ZK_DATADIR + "\ntickTime=" + fakeConfig.ZK_TICKTIME +
				"\nmaxClientCnxns=60\nminSessionTimeout=4000\nmaxSessionTimeout=40000\nserverId=0",
			TestDescription: "Successfully got data",
		},
	}

	for _, test := range tests {
		g.Describe("getFLWconf()", func() {
			g.It(test.TestDescription, func() {
				result := getFLWconf(logrus.New(), fakeConfig)
				fmt.Println(result)
				g.Assert(reflect.DeepEqual(result, test.ExpectedResult)).Equal(true)
			})
		})
	}
}

func TestScrapeFLWconf(t *testing.T) {
	g := goblin.Goblin(t)

	result := map[string]interface{}{
		"event_type":                       "ZookeeperServerSample",
		"provider":                         PROVIDER,
		"zookeeper.conf.clientPort":        confPort,
		"zookeeper.conf.dataDir":           "/var/log/zookeeper",
		"zookeeper.conf.tickTime":          2000,
		"zookeeper.conf.maxClientCnxns":    60,
		"zookeeper.conf.minSessionTimeout": 4000,
		"zookeeper.conf.maxSessionTimeout": 40000,
		"zookeeper.conf.serverId":          0,
	}

	var tests = []struct {
		Data            string
		ExpectedResult  map[string]interface{}
		TestDescription string
	}{
		{
			Data:            "clientPort=" + fakeConfig.ZK_CLIENTPORT + "\ndataDir=/var/log/zookeeper\ntickTime=2000\nmaxClientCnxns=60\nminSessionTimeout=4000\nmaxSessionTimeout=40000\nserverId=0",
			ExpectedResult:  result,
			TestDescription: "Successfully scraped FLW command conf",
		},
	}

	for _, test := range tests {
		g.Describe("ScrapeFLWconf()", func() {
			g.It(test.TestDescription, func() {
				result := ScrapeFLWconf(logrus.New(), test.Data)
				fmt.Println(result)
				g.Assert(reflect.DeepEqual(result, test.ExpectedResult)).Equal(true)
			})
		})
	}
}

func TestScrapeFLWmntr(t *testing.T) {
	g := goblin.Goblin(t)

	result := map[string]interface{}{
		"event_type":                                   "ZookeeperServerSample",
		"provider":                                     PROVIDER,
		"zookeeper.mntr.zk_version":                    "3.4.10-39d3a4f269333c922ed3db283be479f9deacaa0f",
		"zookeeper.mntr.zk_avg_latency":                0,
		"zookeeper.mntr.zk_max_latency":                0,
		"zookeeper.mntr.zk_min_latency":                0,
		"zookeeper.mntr.zk_packets_received":           2,
		"zookeeper.mntr.zk_packets_sent":               1,
		"zookeeper.mntr.zk_num_alive_connections":      1,
		"zookeeper.mntr.zk_outstanding_requests":       0,
		"zookeeper.mntr.zk_server_state":               "standalone",
		"zookeeper.mntr.zk_znode_count":                19,
		"zookeeper.mntr.zk_watch_count":                0,
		"zookeeper.mntr.zk_ephemerals_count":           0,
		"zookeeper.mntr.zk_approximate_data_size":      270,
		"zookeeper.mntr.zk_open_file_descriptor_count": 38,
		"zookeeper.mntr.zk_max_file_descriptor_count":  10240,
	}

	var tests = []struct {
		Data            string
		ExpectedResult  map[string]interface{}
		TestDescription string
	}{
		{
			Data: "zk_version\t3.4.10-39d3a4f269333c922ed3db283be479f9deacaa0f\nzk_avg_latency\t0\nzk_max_latency\t0\nzk_min_latency\t0\nzk_packets_received\t2\n" +
				"zk_packets_sent\t1\nzk_num_alive_connections\t1\nzk_outstanding_requests\t0\nzk_server_state\tstandalone\n" +
				"zk_znode_count\t19\nzk_watch_count\t0\nzk_ephemerals_count\t0\nzk_approximate_data_size\t270\n" +
				"zk_open_file_descriptor_count\t38\nzk_max_file_descriptor_count\t10240\n",
			ExpectedResult:  result,
			TestDescription: "Successfully scraped FLW command mntr",
		},
	}

	for _, test := range tests {
		g.Describe("ScrapeFLWmntr()", func() {
			g.It(test.TestDescription, func() {
				result := ScrapeFLWmntr(logrus.New(), test.Data)
				fmt.Println(result)
				g.Assert(reflect.DeepEqual(result, test.ExpectedResult)).Equal(true)
			})
		})
	}
}

func TestToInt(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		Value           string
		ExpectedResult  int
		TestDescription string
	}{
		{
			Value:           "8675301",
			ExpectedResult:  8675301,
			TestDescription: "Should return int 8675301 of string",
		},
		{
			Value:           "",
			ExpectedResult:  0,
			TestDescription: "Should return 0 if empty string",
		},
		{
			Value:           "arf",
			ExpectedResult:  0,
			TestDescription: "Should return 0 if error converting to int",
		},
	}

	for _, test := range tests {
		g.Describe("toInt()", func() {
			g.It(test.TestDescription, func() {
				result := toInt(logrus.New(), test.Value)
				g.Assert(reflect.DeepEqual(result, test.ExpectedResult)).Equal(true)
			})
		})
	}
}
