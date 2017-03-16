package redis

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	redis "gopkg.in/redis.v5"

	"github.com/Sirupsen/logrus"
)

// NAME - name of plugin
const NAME string = "redis"

// PROVIDER -
const PROVIDER string = "redis"

// EVENTTYPE -
const EVENTTYPE string = "RedisInfo"

// ProtocolVersion -
const ProtocolVersion string = "1"

// Config is the keeper of the config
type Config struct {
	RedisHost string // Optional: leaving blank will default to localhost
	RedisPort string // Optional: leaving blank will default to 6379
	RedisPass string // Optional: leaving blank means no password
	RedisDB   string // Optional: leaving blank will keep DBID at 0
	DBID      int    // Not from external config, but holder for DBID int value if specified
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
	ProtocolVersion string                   `json:"ProtocolVersion"`
	PluginVersion   string                   `json:"plugin_version"`
	Metrics         []MetricData             `json:"metrics"`
	Inventory       map[string]InventoryData `json:"inventory"`
	Events          []EventData              `json:"events"`
	Status          string                   `json:"status"`
}

// Client objec to use in communication with redis
var redisClient *redis.Client

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

// Run -
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

	var redisConf = Config{
		RedisHost: os.Getenv("REDISHOST"),
		RedisPort: os.Getenv("REDISPORT"),
		RedisPass: os.Getenv("REDISPASS"),
		RedisDB:   os.Getenv("REDISDB"),
	}
	validateConfig(log, &redisConf)

	initRedisClient(redisConf)

	var metric = formatMetric(log, readStats(log, redisConf))

	data.Metrics = append(data.Metrics, metric)
	fatalIfErr(log, OutputJSON(data, prettyPrint))
}

func initRedisClient(conf Config) {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", conf.RedisHost, conf.RedisPort),
		Password: conf.RedisPass,
		DB:       conf.DBID,
	})
}

func fatalIfErr(log *logrus.Logger, err error) {
	if err != nil {
		log.WithError(err).Fatal("can't continue")
	}
}

func validateConfig(log *logrus.Logger, redisConf *Config) {
	if redisConf.RedisHost == "" {
		redisConf.RedisHost = "localhost"
	}

	if redisConf.RedisPort == "" {
		redisConf.RedisPort = "6379"
	} else {
		_, err := strconv.Atoi(redisConf.RedisPort)
		if err != nil {
			log.WithError(err).Fatal("Config Yaml value REDISPORT must be valid integer")
		}
	}

	if redisConf.RedisDB != "" {
		val, err := strconv.Atoi(redisConf.RedisDB)
		if err != nil {
			log.WithError(err).Fatal("Config Yaml value REDISDB must be valid integer")
		}
		redisConf.DBID = val
	}
}

func readStats(log *logrus.Logger, redisConf Config) string {
	output, err := redisClient.Info().Result()
	if err != nil {
		log.WithError(err).Fatal("Error making stats call to redis")
	}
	return output
}

func parseRawData(rawMetric string) map[string]string {
	results := map[string]string{}
	lines := strings.Split(rawMetric, "\n")
	for _, line := range lines {
		splitLine := strings.Split(line, ":")
		if len(splitLine) == 2 {
			results[splitLine[0]] = strings.Trim(splitLine[1], "\r")
		}
	}
	return results
}

func formatMetric(log *logrus.Logger, rawMetric string) map[string]interface{} {
	log.WithFields(logrus.Fields{
		"output": rawMetric,
	}).Debugf("Full raw info response text")
	rawData := parseRawData(rawMetric)

	return map[string]interface{}{
		"event_type":                           EVENTTYPE,
		"providor":                             PROVIDER,
		"redis.redis_version":                  rawData["redis_version"],
		"redis.redis_git_sha1":                 rawData["redis_git_sha1"],
		"redis.redis_git_dirty":                toInt(log, rawData["redis_git_dirty"]),
		"redis.redis_build_id":                 rawData["redis_build_id"],
		"redis.redis_mode":                     rawData["redis_mode"],
		"redis.os":                             rawData["os"],
		"redis.arch_bits":                      toInt(log, rawData["arch_bits"]),
		"redis.multiplexing_api":               rawData["multiplexing_api"],
		"redis.gcc_version":                    rawData["gcc_version"],
		"redis.process_id":                     toInt(log, rawData["process_id"]),
		"redis.run_id":                         rawData["run_id"],
		"redis.tcp_port":                       toInt(log, rawData["tcp_port"]),
		"redis.uptime_in_seconds":              toInt(log, rawData["uptime_in_seconds"]),
		"redis.uptime_in_days":                 toInt(log, rawData["uptime_in_days"]),
		"redis.hz":                             toInt(log, rawData["hz"]),
		"redis.lru_clock":                      toInt(log, rawData["lru_clock"]),
		"redis.executable":                     rawData["executable"],
		"redis.config_file":                    rawData["config_file"],
		"redis.connected_clients":              toInt(log, rawData["connected_clients"]),
		"redis.client_longest_output_list":     toInt(log, rawData["client_longest_output_list"]),
		"redis.client_biggest_input_buf":       toInt(log, rawData["client_biggest_input_buf"]),
		"redis.blocked_clients":                toInt(log, rawData["blocked_clients"]),
		"redis.used_memory":                    toInt(log, rawData["used_memory"]),
		"redis.used_memory_human":              rawData["used_memory_human"],
		"redis.used_memory_rss":                toInt(log, rawData["used_memory_rss"]),
		"redis.used_memory_rss_human":          rawData["used_memory_rss_human"],
		"redis.used_memory_peak":               toInt(log, rawData["used_memory_peak"]),
		"redis.used_memory_peak_human":         rawData["used_memory_peak_human"],
		"redis.total_system_memory":            toInt(log, rawData["total_system_memory"]),
		"redis.total_system_memory_human":      rawData["total_system_memory_human"],
		"redis.used_memory_lua":                toInt(log, rawData["used_memory_lua"]),
		"redis.used_memory_lua_human":          rawData["used_memory_lua_human"],
		"redis.maxmemory":                      toInt(log, rawData["maxmemory"]),
		"redis.maxmemory_human":                rawData["maxmemory_human"],
		"redis.maxmemory_policy":               rawData["maxmemory_policy"],
		"redis.mem_fragmentation_ratio":        toFloat(log, rawData["mem_fragmentation_ratio"]),
		"redis.mem_allocator":                  rawData["mem_allocator"],
		"redis.loading":                        toInt(log, rawData["loading"]),
		"redis.rdb_changes_since_last_save":    toInt(log, rawData["rdb_changes_since_last_save"]),
		"redis.rdb_bgsave_in_progress":         toInt(log, rawData["rdb_bgsave_in_progress"]),
		"redis.rdb_last_save_time":             toInt(log, rawData["rdb_last_save_time"]),
		"redis.rdb_last_bgsave_status":         rawData["rdb_last_bgsave_status"],
		"redis.rdb_last_bgsave_time_sec":       toInt(log, rawData["rdb_last_bgsave_time_sec"]),
		"redis.rdb_current_bgsave_time_sec":    toInt(log, rawData["rdb_current_bgsave_time_sec"]),
		"redis.aof_enabled":                    toInt(log, rawData["aof_enabled"]),
		"redis.aof_rewrite_in_progress":        toInt(log, rawData["aof_rewrite_in_progress"]),
		"redis.aof_rewrite_scheduled":          toInt(log, rawData["aof_rewrite_scheduled"]),
		"redis.aof_last_rewrite_time_sec":      toInt(log, rawData["aof_last_rewrite_time_sec"]),
		"redis.aof_current_rewrite_time_sec":   toInt(log, rawData["aof_current_rewrite_time_sec"]),
		"redis.aof_last_bgrewrite_status":      rawData["aof_last_bgrewrite_status"],
		"redis.aof_last_write_status":          rawData["aof_last_write_status"],
		"redis.total_connections_received":     toInt(log, rawData["total_connections_received"]),
		"redis.total_commands_processed":       toInt(log, rawData["total_commands_processed"]),
		"redis.instantaneous_ops_per_sec":      toInt(log, rawData["instantaneous_ops_per_sec"]),
		"redis.total_net_input_bytes":          toInt(log, rawData["total_net_input_bytes"]),
		"redis.total_net_output_bytes":         toInt(log, rawData["total_net_output_bytes"]),
		"redis.instantaneous_input_kbps":       toFloat(log, rawData["instantaneous_input_kbps"]),
		"redis.instantaneous_output_kbps":      toFloat(log, rawData["instantaneous_output_kbps"]),
		"redis.rejected_connections":           toInt(log, rawData["rejected_connections"]),
		"redis.sync_full":                      toInt(log, rawData["sync_full"]),
		"redis.sync_partial_ok":                toInt(log, rawData["sync_partial_ok"]),
		"redis.sync_partial_err":               toInt(log, rawData["sync_partial_err"]),
		"redis.expired_keys":                   toInt(log, rawData["expired_keys"]),
		"redis.evicted_keys":                   toInt(log, rawData["evicted_keys"]),
		"redis.keyspace_hits":                  toInt(log, rawData["keyspace_hits"]),
		"redis.keyspace_misses":                toInt(log, rawData["keyspace_misses"]),
		"redis.pubsub_channels":                toInt(log, rawData["pubsub_channels"]),
		"redis.pubsub_patterns":                toInt(log, rawData["pubsub_patterns"]),
		"redis.latest_fork_usec":               toInt(log, rawData["latest_fork_usec"]),
		"redis.migrate_cached_sockets":         toInt(log, rawData["migrate_cached_sockets"]),
		"redis.role":                           rawData["role"],
		"redis.connected_slaves":               toInt(log, rawData["connected_slaves"]),
		"redis.master_repl_offset":             toInt(log, rawData["master_repl_offset"]),
		"redis.repl_backlog_active":            toInt(log, rawData["repl_backlog_active"]),
		"redis.repl_backlog_size":              toInt(log, rawData["repl_backlog_size"]),
		"redis.repl_backlog_first_byte_offset": toInt(log, rawData["repl_backlog_first_byte_offset"]),
		"redis.repl_backlog_histlen":           toInt(log, rawData["repl_backlog_histlen"]),
		"redis.used_cpu_sys":                   toFloat(log, rawData["used_cpu_sys"]),
		"redis.used_cpu_user":                  toFloat(log, rawData["used_cpu_user"]),
		"redis.used_cpu_sys_children":          toFloat(log, rawData["used_cpu_sys_children"]),
		"redis.used_cpu_user_children":         toFloat(log, rawData["used_cpu_user_children"]),
		"redis.cluster_enabled":                toInt(log, rawData["cluster_enabled"]),
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

func toFloat(log *logrus.Logger, value string) float64 {
	if value == "" {
		return 0.00
	}
	valueFloat, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.WithFields(logrus.Fields{
			"valueFloat": valueFloat,
			"error":      err,
		}).Debug("Error converting value to float")

		return 0.00
	}

	return valueFloat
}
