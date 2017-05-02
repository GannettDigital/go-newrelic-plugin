package zookeeper

import (
	"os"
	"strconv"
	"strings"
	"github.com/Sirupsen/logrus"
	"time"
	"net"
	"io/ioutil"
	"github.com/GannettDigital/go-newrelic-plugin/helpers"
)

// NAME - name of plugin
const NAME string = "zookeeper"

// PROVIDER -
const PROVIDER string = "zookeeper" 

// ProtocolVersion -
const ProtocolVersion string = "1"

//Config is the keeper of the config
type Config struct {
	ZK_TICKTIME 		string
	ZK_DATADIR  		string
	ZK_HOST     		string
	ZK_CLIENTPORT     	string
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

	var ZKConf = Config{
		ZK_TICKTIME:  	os.Getenv("ZK_TICKTIME"),
		ZK_DATADIR:     os.Getenv("ZK_DATADIR"),
		ZK_HOST:  	os.Getenv("ZK_HOST"),
		ZK_CLIENTPORT:  os.Getenv("ZK_CLIENTPORT"),
	}

	validateConfig(log, ZKConf)

	var conf_metric = ScrapeFLWconf(log, getFLWconf(log, ZKConf))
	data.Metrics = append(data.Metrics, conf_metric)

	var mntr_metric = ScrapeFLWmntr(log, getFLWmntr(log, ZKConf))
	data.Metrics = append(data.Metrics, mntr_metric)

	fatalIfErr(log, helpers.OutputJSON(data, prettyPrint))
}

func validateConfig(log *logrus.Logger, ZKConf Config)  {
	if ZKConf.ZK_TICKTIME == "" {
		log.Fatal("Config is missing the ZK_TICKTIME. Please check the config to continue")
	}
	if ZKConf.ZK_DATADIR == "" {
		log.Fatal("Config is missing the ZK_DATADIR. Please check the config to continue")
	}
	if ZKConf.ZK_HOST == "" {
		log.Fatal("Config is missing the ZK_HOST. Please check the config to continue")

	}
	if ZKConf.ZK_CLIENTPORT == "" {
		log.Fatal("Config is missing the ZK_CLIENTPORT. Please check the config to continue")

	}
}


func fatalIfErr(log *logrus.Logger, err error) {
	if err != nil {
		log.WithError(err).Fatal("can't continue")
	}
}

func getFLWconf(log *logrus.Logger, ZKConf Config) string {

	tickTime, _ := strconv.Atoi(ZKConf.ZK_TICKTIME)
	timeOut := time.Duration(tickTime) * time.Millisecond

	conn, err := net.DialTimeout("tcp", ZKConf.ZK_HOST + ":" + ZKConf.ZK_CLIENTPORT, timeOut)
	fatalIfErr(log, err)

	if err != nil {
		log.WithFields(logrus.Fields{
			"config.zookeeper.ZK_HOST": ZKConf.ZK_HOST,
			"config.zookeeper.ZK_CLIENTPORT": ZKConf.ZK_CLIENTPORT,
			"config.zookeeper.ZK_TICKTIME": ZKConf.ZK_TICKTIME,
			"config.zookeeper.ZK_DATADIR": ZKConf.ZK_DATADIR,
			"error":                  err,
		}).Fatal("Encountered error calling conf")
		return ""
	}

	// close the connection
	defer conn.Close()

	//Read status using conf
	_, err = conn.Write([]byte("conf"))
	fatalIfErr(log, err)

	conn.SetReadDeadline(time.Now().Add(time.Duration(timeOut)))

	everything, err := ioutil.ReadAll(conn)
	return string(everything)
}

func getFLWmntr(log *logrus.Logger, ZKConf Config) string {

	tickTime, _ := strconv.Atoi(ZKConf.ZK_TICKTIME)
	timeOut := time.Duration(tickTime) * time.Millisecond

	conn, err := net.DialTimeout("tcp", ZKConf.ZK_HOST + ":" + ZKConf.ZK_CLIENTPORT, timeOut)
	fatalIfErr(log, err)

	if err != nil {
		log.WithFields(logrus.Fields{
			"config.zookeeper.ZK_HOST": ZKConf.ZK_HOST,
			"config.zookeeper.ZK_CLIENTPORT": ZKConf.ZK_CLIENTPORT,
			"config.zookeeper.ZK_TICKTIME": ZKConf.ZK_TICKTIME,
			"config.zookeeper.ZK_DATADIR": ZKConf.ZK_DATADIR,
			"error":                  err,
		}).Fatal("Encountered error calling mntr")
		return ""
	}

	// close the connection
	defer conn.Close()

	//Read status using mntr
	_, err = conn.Write([]byte("mntr"))
	fatalIfErr(log, err)

	conn.SetReadDeadline(time.Now().Add(time.Duration(timeOut)))

	everything, err := ioutil.ReadAll(conn)
	return string(everything)

}

func ScrapeFLWconf(log *logrus.Logger, status string) map[string]interface{} {
	line := strings.Split(string(status), "\n")

	clientPort := strings.Split(string(line[0]), "=")
	dataDir := strings.Split(string(line[1]), "=")
	tickTime := strings.Split(string(line[2]), "=")
	maxClientCnxns := strings.Split(string(line[3]), "=")
	minSessionTimeout := strings.Split(string(line[4]), "=")
	maxSessionTimeout := strings.Split(string(line[5]), "=")
	serverId := strings.Split(string(line[6]), "=")

	log.WithFields(logrus.Fields{
		"clientPort":  clientPort[1],
		"dataDir":   dataDir[1],
		"tickTime":  tickTime[1],
		"maxClientCnxns":  maxClientCnxns[1],
		"minSessionTimeout":  minSessionTimeout[1],
		"maxSessionTimeout":  maxSessionTimeout[1],
		"serverId":  serverId[1],
	}).Debugf("Scraped ZooKeeper conf values")
	return map[string]interface{}{
		"event_type":            "ZookeeperServerSample",
		"provider":              PROVIDER,
		"zookeeper.conf.clientPort":     toInt(log, clientPort[1]),
		"zookeeper.conf.dataDir": dataDir[1],
		"zookeeper.conf.tickTime": toInt(log, tickTime[1]),
		"zookeeper.conf.maxClientCnxns": toInt(log, maxClientCnxns[1]),
		"zookeeper.conf.minSessionTimeout": toInt(log, minSessionTimeout[1]),
		"zookeeper.conf.maxSessionTimeout": toInt(log, maxSessionTimeout[1]),
		"zookeeper.conf.serverId": toInt(log, serverId[1]),
	}
}

func ScrapeFLWmntr(log *logrus.Logger, status string) map[string]interface{} {
	line := strings.Split(string(status), "\n")

	zk_version := strings.Split(string(line[0]), "\t")
	zk_avg_latency := strings.Split(string(line[1]), "\t")
	zk_max_latency := strings.Split(string(line[2]), "\t")
	zk_min_latency := strings.Split(string(line[3]), "\t")
	zk_packets_received := strings.Split(string(line[4]), "\t")
	zk_packets_sent := strings.Split(string(line[5]), "\t")
	zk_num_alive_connections := strings.Split(string(line[6]), "\t")
	zk_outstanding_requests := strings.Split(string(line[7]), "\t")
	zk_server_state := strings.Split(string(line[8]), "\t")
	zk_znode_count := strings.Split(string(line[9]), "\t")
	zk_watch_count := strings.Split(string(line[10]), "\t")
	zk_ephemerals_count := strings.Split(string(line[11]), "\t")
	zk_approximate_data_size := strings.Split(string(line[12]), "\t")
	zk_open_file_descriptor_count := strings.Split(string(line[13]), "\t")
	zk_max_file_descriptor_count := strings.Split(string(line[14]), "\t")

	// Add additional attributes if this server is the leader
	if (zk_server_state[1] == "leader") {

		zk_followers := strings.Split(string(line[15]), "\t")
		zk_synced_followers := strings.Split(string(line[14]), "\t")
		zk_pending_syncs := strings.Split(string(line[14]), "\t")

		log.WithFields(logrus.Fields{
			"zk_version":  zk_version[1],
			"zk_avg_latency":   zk_avg_latency[1],
			"zk_max_latency":  zk_max_latency[1],
			"zk_min_latency":  zk_min_latency[1],
			"zk_packets_received":  zk_packets_received[1],
			"zk_packets_sent":  zk_packets_sent[1],
			"zk_num_alive_connections":  zk_num_alive_connections[1],
			"zk_outstanding_requests":  zk_outstanding_requests[1],
			"zk_server_state":  zk_server_state[1],
			"zk_znode_count":  zk_znode_count[1],
			"zk_watch_count":  zk_watch_count[1],
			"zk_ephemerals_count":  zk_ephemerals_count[1],
			"zk_approximate_data_size":  zk_approximate_data_size[1],
			"zk_open_file_descriptor_count":  zk_open_file_descriptor_count[1],
			"zk_max_file_descriptor_count":  zk_max_file_descriptor_count[1],
			"zk_followers":  zk_followers[1],
			"zk_synced_followers":  zk_synced_followers[1],
			"zk_pending_syncs":  zk_pending_syncs[1],
		}).Debugf("Scraped ZooKeeper values")
		return map[string]interface{}{
			"event_type":            "ZookeeperServerSample",
			"provider":              PROVIDER,
			"zookeeper.mntr.zk_version":     zk_version[1],
			"zookeeper.mntr.zk_avg_latency": toInt(log, zk_avg_latency[1]),
			"zookeeper.mntr.zk_max_latency": toInt(log, zk_max_latency[1]),
			"zookeeper.mntr.zk_min_latency": toInt(log, zk_min_latency[1]),
			"zookeeper.mntr.zk_packets_received": toInt(log, zk_packets_received[1]),
			"zookeeper.mntr.zk_packets_sent": toInt(log, zk_packets_sent[1]),
			"zookeeper.mntr.zk_num_alive_connections": toInt(log, zk_num_alive_connections[1]),
			"zookeeper.mntr.zk_outstanding_requests": toInt(log, zk_outstanding_requests[1]),
			"zookeeper.mntr.zk_server_state": zk_server_state[1],
			"zookeeper.mntr.zk_znode_count": toInt(log, zk_znode_count[1]),
			"zookeeper.mntr.zk_watch_count": toInt(log, zk_watch_count[1]),
			"zookeeper.mntr.zk_ephemerals_count": toInt(log, zk_ephemerals_count[1]),
			"zookeeper.mntr.zk_approximate_data_size": toInt(log, zk_approximate_data_size[1]),
			"zookeeper.mntr.zk_open_file_descriptor_count": toInt(log, zk_open_file_descriptor_count[1]),
			"zookeeper.mntr.zk_max_file_descriptor_count": toInt(log, zk_max_file_descriptor_count[1]),
			"zookeeper.mntr.zk_followers": toInt(log, zk_followers[1]),
			"zookeeper.mntr.zk_synced_followers": toInt(log, zk_synced_followers[1]),
			"zookeeper.mntr.zk_pending_syncs": toInt(log, zk_pending_syncs[1]),

		}
	}

	log.WithFields(logrus.Fields{
		"zk_version":  zk_version[1],
		"zk_avg_latency":   zk_avg_latency[1],
		"zk_max_latency":  zk_max_latency[1],
		"zk_min_latency":  zk_min_latency[1],
		"zk_packets_received":  zk_packets_received[1],
		"zk_packets_sent":  zk_packets_sent[1],
		"zk_num_alive_connections":  zk_num_alive_connections[1],
		"zk_outstanding_requests":  zk_outstanding_requests[1],
		"zk_server_state":  zk_server_state[1],
		"zk_znode_count":  zk_znode_count[1],
		"zk_watch_count":  zk_watch_count[1],
		"zk_ephemerals_count":  zk_ephemerals_count[1],
		"zk_approximate_data_size":  zk_approximate_data_size[1],
		"zk_open_file_descriptor_count":  zk_open_file_descriptor_count[1],
		"zk_max_file_descriptor_count":  zk_max_file_descriptor_count[1],
	}).Debugf("Scraped ZooKeeper values")
	return map[string]interface{}{
		"event_type":            "ZookeeperServerSample",
		"provider":              PROVIDER,
		"zookeeper.mntr.zk_version":     zk_version[1],
		"zookeeper.mntr.zk_avg_latency": toInt(log, zk_avg_latency[1]),
		"zookeeper.mntr.zk_max_latency": toInt(log, zk_max_latency[1]),
		"zookeeper.mntr.zk_min_latency": toInt(log, zk_min_latency[1]),
		"zookeeper.mntr.zk_packets_received": toInt(log, zk_packets_received[1]),
		"zookeeper.mntr.zk_packets_sent": toInt(log, zk_packets_sent[1]),
		"zookeeper.mntr.zk_num_alive_connections": toInt(log, zk_num_alive_connections[1]),
		"zookeeper.mntr.zk_outstanding_requests": toInt(log, zk_outstanding_requests[1]),
		"zookeeper.mntr.zk_server_state": zk_server_state[1],
		"zookeeper.mntr.zk_znode_count": toInt(log, zk_znode_count[1]),
		"zookeeper.mntr.zk_watch_count": toInt(log, zk_watch_count[1]),
		"zookeeper.mntr.zk_ephemerals_count": toInt(log, zk_ephemerals_count[1]),
		"zookeeper.mntr.zk_approximate_data_size": toInt(log, zk_approximate_data_size[1]),
		"zookeeper.mntr.zk_open_file_descriptor_count": toInt(log, zk_open_file_descriptor_count[1]),
		"zookeeper.mntr.zk_max_file_descriptor_count": toInt(log, zk_max_file_descriptor_count[1]),
	}
}

func toInt(log *logrus.Logger, value string) int {
	if value == "" {
		return 0
	}
	valueInt, err := strconv.Atoi(value)
	if err != nil {
		log.WithFields(logrus.Fields{
			"valueInt": valueInt,
			"error":    err,
		}).Debug("Error converting value to int")

		return 0
	}

	return valueInt
}