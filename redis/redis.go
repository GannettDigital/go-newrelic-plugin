package redis

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	redis "gopkg.in/redis.v5"

	"github.com/Sirupsen/logrus"
	"github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/metric"
	newrelicsdk "github.com/newrelic/infra-integrations-sdk/sdk"
)

// NAME - name of plugin
const NAME string = "redis"

// PROVIDER -
const PROVIDER string = "redis"

// EVENTTYPE -
const EVENTTYPE string = "RedisInfo"

// ProtocolVersion -
const ProtocolVersion string = "1"

// RedisClientImpl - interface used for mocking
type RedisClientImpl interface {
	Info(section ...string) *redis.StringCmd
}

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
	ProtocolVersion string                   `json:"protocol_version"`
	PluginVersion   string                   `json:"plugin_version"`
	Metrics         []MetricData             `json:"metrics"`
	Inventory       map[string]InventoryData `json:"inventory"`
	Events          []EventData              `json:"events"`
	Status          string                   `json:"status"`
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

// Run -
func Run(log *logrus.Logger, client RedisClientImpl, redisConf Config, prettyPrint bool, version string) {
	integration, err := newrelicsdk.NewIntegration(NAME, version, &args.DefaultArgumentList{})
	fatalIfErr(log, err)

	formatMetrics(log, integration, readStats(log, client, redisConf))
	fatalIfErr(log, integration.Publish())
}

// InitRedisClient - function to create a redis client
func InitRedisClient(conf Config) RedisClientImpl {
	return redis.NewClient(&redis.Options{
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

// ValidateConfig - function to validate the config and set defaults
func ValidateConfig(log *logrus.Logger, redisConf *Config) {
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

func readStats(log *logrus.Logger, client RedisClientImpl, redisConf Config) string {
	output, err := client.Info().Result()
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

func formatMetrics(log *logrus.Logger, integration *newrelicsdk.Integration, rawMetric string) {
	log.WithFields(logrus.Fields{
		"output": rawMetric,
	}).Debugf("Full raw info response text")
	rawData := parseRawData(rawMetric)

	metricSet := integration.NewMetricSet(EVENTTYPE)
	metricSet.SetMetric("providor", PROVIDER, metric.ATTRIBUTE)
	metricSet.SetMetric("redis.redis_version", rawData["redis_version"], metric.ATTRIBUTE)
	metricSet.SetMetric("redis.redis_git_sha1", rawData["redis_git_sha1"], metric.ATTRIBUTE)
	metricSet.SetMetric("redis.redis_git_dirty", toInt(log, rawData["redis_git_dirty"]), metric.GAUGE)
	metricSet.SetMetric("redis.redis_build_id", rawData["redis_build_id"], metric.ATTRIBUTE)
	metricSet.SetMetric("redis.redis_mode", rawData["redis_mode"], metric.ATTRIBUTE)
	metricSet.SetMetric("redis.os", rawData["os"], metric.ATTRIBUTE)
	metricSet.SetMetric("redis.arch_bits", toInt(log, rawData["arch_bits"]), metric.GAUGE)
	metricSet.SetMetric("redis.multiplexing_api", rawData["multiplexing_api"], metric.ATTRIBUTE)
	metricSet.SetMetric("redis.gcc_version", rawData["gcc_version"], metric.ATTRIBUTE)
	metricSet.SetMetric("redis.process_id", toInt(log, rawData["process_id"]), metric.GAUGE)
	metricSet.SetMetric("redis.run_id", rawData["run_id"], metric.ATTRIBUTE)
	metricSet.SetMetric("redis.tcp_port", toInt(log, rawData["tcp_port"]), metric.GAUGE)
	metricSet.SetMetric("redis.uptime_in_seconds", toInt(log, rawData["uptime_in_seconds"]), metric.GAUGE)
	metricSet.SetMetric("redis.uptime_in_days", toInt(log, rawData["uptime_in_days"]), metric.GAUGE)
	metricSet.SetMetric("redis.hz", toInt(log, rawData["hz"]), metric.GAUGE)
	metricSet.SetMetric("redis.lru_clock", toInt(log, rawData["lru_clock"]), metric.GAUGE)
	metricSet.SetMetric("redis.executable", rawData["executable"], metric.ATTRIBUTE)
	metricSet.SetMetric("redis.config_file", rawData["config_file"], metric.ATTRIBUTE)
	metricSet.SetMetric("redis.connected_clients", toInt(log, rawData["connected_clients"]), metric.GAUGE)
	metricSet.SetMetric("redis.client_longest_output_list", toInt(log, rawData["client_longest_output_list"]), metric.GAUGE)
	metricSet.SetMetric("redis.client_biggest_input_buf", toInt(log, rawData["client_biggest_input_buf"]), metric.GAUGE)
	metricSet.SetMetric("redis.blocked_clients", toInt(log, rawData["blocked_clients"]), metric.GAUGE)
	metricSet.SetMetric("redis.used_memory", toInt(log, rawData["used_memory"]), metric.GAUGE)
	metricSet.SetMetric("redis.used_memory_human", rawData["used_memory_human"], metric.ATTRIBUTE)
	metricSet.SetMetric("redis.used_memory_rss", toInt(log, rawData["used_memory_rss"]), metric.GAUGE)
	metricSet.SetMetric("redis.used_memory_rss_human", rawData["used_memory_rss_human"], metric.ATTRIBUTE)
	metricSet.SetMetric("redis.used_memory_peak", toInt(log, rawData["used_memory_peak"]), metric.GAUGE)
	metricSet.SetMetric("redis.used_memory_peak_human", rawData["used_memory_peak_human"], metric.ATTRIBUTE)
	metricSet.SetMetric("redis.total_system_memory", toInt(log, rawData["total_system_memory"]), metric.GAUGE)
	metricSet.SetMetric("redis.total_system_memory_human", rawData["total_system_memory_human"], metric.ATTRIBUTE)
	metricSet.SetMetric("redis.used_memory_lua", toInt(log, rawData["used_memory_lua"]), metric.GAUGE)
	metricSet.SetMetric("redis.used_memory_lua_human", rawData["used_memory_lua_human"], metric.ATTRIBUTE)
	metricSet.SetMetric("redis.maxmemory", toInt(log, rawData["maxmemory"]), metric.GAUGE)
	metricSet.SetMetric("redis.maxmemory_human", rawData["maxmemory_human"], metric.ATTRIBUTE)
	metricSet.SetMetric("redis.maxmemory_policy", rawData["maxmemory_policy"], metric.ATTRIBUTE)
	metricSet.SetMetric("redis.mem_fragmentation_ratio", toFloat(log, rawData["mem_fragmentation_ratio"]), metric.GAUGE)
	metricSet.SetMetric("redis.mem_allocator", rawData["mem_allocator"], metric.ATTRIBUTE)
	metricSet.SetMetric("redis.loading", toInt(log, rawData["loading"]), metric.GAUGE)
	metricSet.SetMetric("redis.rdb_changes_since_last_save", toInt(log, rawData["rdb_changes_since_last_save"]), metric.GAUGE)
	metricSet.SetMetric("redis.rdb_bgsave_in_progress", toInt(log, rawData["rdb_bgsave_in_progress"]), metric.GAUGE)
	metricSet.SetMetric("redis.rdb_last_save_time", toInt(log, rawData["rdb_last_save_time"]), metric.GAUGE)
	metricSet.SetMetric("redis.rdb_last_bgsave_status", rawData["rdb_last_bgsave_status"], metric.ATTRIBUTE)
	metricSet.SetMetric("redis.rdb_last_bgsave_time_sec", toInt(log, rawData["rdb_last_bgsave_time_sec"]), metric.GAUGE)
	metricSet.SetMetric("redis.rdb_current_bgsave_time_sec", toInt(log, rawData["rdb_current_bgsave_time_sec"]), metric.GAUGE)
	metricSet.SetMetric("redis.aof_enabled", toInt(log, rawData["aof_enabled"]), metric.GAUGE)
	metricSet.SetMetric("redis.aof_rewrite_in_progress", toInt(log, rawData["aof_rewrite_in_progress"]), metric.GAUGE)
	metricSet.SetMetric("redis.aof_rewrite_scheduled", toInt(log, rawData["aof_rewrite_scheduled"]), metric.GAUGE)
	metricSet.SetMetric("redis.aof_last_rewrite_time_sec", toInt(log, rawData["aof_last_rewrite_time_sec"]), metric.GAUGE)
	metricSet.SetMetric("redis.aof_current_rewrite_time_sec", toInt(log, rawData["aof_current_rewrite_time_sec"]), metric.GAUGE)
	metricSet.SetMetric("redis.aof_last_bgrewrite_status", rawData["aof_last_bgrewrite_status"], metric.ATTRIBUTE)
	metricSet.SetMetric("redis.aof_last_write_status", rawData["aof_last_write_status"], metric.ATTRIBUTE)
	metricSet.SetMetric("redis.total_connections_received", toInt(log, rawData["total_connections_received"]), metric.GAUGE)
	metricSet.SetMetric("redis.total_commands_processed", toInt(log, rawData["total_commands_processed"]), metric.GAUGE)
	metricSet.SetMetric("redis.instantaneous_ops_per_sec", toInt(log, rawData["instantaneous_ops_per_sec"]), metric.GAUGE)
	metricSet.SetMetric("redis.total_net_input_bytes", toInt(log, rawData["total_net_input_bytes"]), metric.GAUGE)
	metricSet.SetMetric("redis.total_net_output_bytes", toInt(log, rawData["total_net_output_bytes"]), metric.GAUGE)
	metricSet.SetMetric("redis.instantaneous_input_kbps", toFloat(log, rawData["instantaneous_input_kbps"]), metric.GAUGE)
	metricSet.SetMetric("redis.instantaneous_output_kbps", toFloat(log, rawData["instantaneous_output_kbps"]), metric.GAUGE)
	metricSet.SetMetric("redis.rejected_connections", toInt(log, rawData["rejected_connections"]), metric.GAUGE)
	metricSet.SetMetric("redis.sync_full", toInt(log, rawData["sync_full"]), metric.GAUGE)
	metricSet.SetMetric("redis.sync_partial_ok", toInt(log, rawData["sync_partial_ok"]), metric.GAUGE)
	metricSet.SetMetric("redis.sync_partial_err", toInt(log, rawData["sync_partial_err"]), metric.GAUGE)
	metricSet.SetMetric("redis.expired_keys", toInt(log, rawData["expired_keys"]), metric.GAUGE)
	metricSet.SetMetric("redis.evicted_keys", toInt(log, rawData["evicted_keys"]), metric.GAUGE)
	metricSet.SetMetric("redis.keyspace_hits", toInt(log, rawData["keyspace_hits"]), metric.GAUGE)
	metricSet.SetMetric("redis.keyspace_misses", toInt(log, rawData["keyspace_misses"]), metric.GAUGE)
	metricSet.SetMetric("redis.pubsub_channels", toInt(log, rawData["pubsub_channels"]), metric.GAUGE)
	metricSet.SetMetric("redis.pubsub_patterns", toInt(log, rawData["pubsub_patterns"]), metric.GAUGE)
	metricSet.SetMetric("redis.latest_fork_usec", toInt(log, rawData["latest_fork_usec"]), metric.GAUGE)
	metricSet.SetMetric("redis.migrate_cached_sockets", toInt(log, rawData["migrate_cached_sockets"]), metric.GAUGE)
	metricSet.SetMetric("redis.role", rawData["role"], metric.ATTRIBUTE)
	metricSet.SetMetric("redis.connected_slaves", toInt(log, rawData["connected_slaves"]), metric.GAUGE)
	metricSet.SetMetric("redis.master_repl_offset", toInt(log, rawData["master_repl_offset"]), metric.GAUGE)
	metricSet.SetMetric("redis.repl_backlog_active", toInt(log, rawData["repl_backlog_active"]), metric.GAUGE)
	metricSet.SetMetric("redis.repl_backlog_size", toInt(log, rawData["repl_backlog_size"]), metric.GAUGE)
	metricSet.SetMetric("redis.repl_backlog_first_byte_offset", toInt(log, rawData["repl_backlog_first_byte_offset"]), metric.GAUGE)
	metricSet.SetMetric("redis.repl_backlog_histlen", toInt(log, rawData["repl_backlog_histlen"]), metric.GAUGE)
	metricSet.SetMetric("redis.used_cpu_sys", toFloat(log, rawData["used_cpu_sys"]), metric.GAUGE)
	metricSet.SetMetric("redis.used_cpu_user", toFloat(log, rawData["used_cpu_user"]), metric.GAUGE)
	metricSet.SetMetric("redis.used_cpu_sys_children", toFloat(log, rawData["used_cpu_sys_children"]), metric.GAUGE)
	metricSet.SetMetric("redis.used_cpu_user_children", toFloat(log, rawData["used_cpu_user_children"]), metric.GAUGE)
	metricSet.SetMetric("redis.cluster_enabled", toInt(log, rawData["cluster_enabled"]), metric.GAUGE)
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
