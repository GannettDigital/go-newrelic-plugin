package redis

import (
	"encoding/json"
	"testing"

	redis "gopkg.in/redis.v5"

	"github.com/GannettDigital/go-newrelic-plugin/redis/fake"
	"github.com/Sirupsen/logrus"
	"github.com/franela/goblin"
)

func TestOutputJSON(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputData       interface{}
		InputPretty     bool
		ExpectedErr     error
		TestDescription string
	}{
		{
			InputData: map[string]interface{}{
				"thing": "stuff",
			},
			InputPretty:     false,
			ExpectedErr:     nil,
			TestDescription: "Should return no error with valid input and pretty of false",
		},
		{
			InputData: map[string]interface{}{
				"thing": "stuff",
			},
			InputPretty:     true,
			ExpectedErr:     nil,
			TestDescription: "Should return no error with valid input and pretty of true",
		},
		{
			InputData:       nil,
			InputPretty:     false,
			ExpectedErr:     nil,
			TestDescription: "Should return no error when nil value is provided",
		},
	}

	for _, test := range tests {
		g.Describe("OutputJSON()", func() {
			g.It(test.TestDescription, func() {
				err := OutputJSON(test.InputData, test.InputPretty)
				g.Assert(err).Equal(test.ExpectedErr)
			})
		})
	}
}

func TestRun(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputLog        *logrus.Logger
		InputClient     RedisClientImpl
		InputConfig     Config
		InputPretty     bool
		InputVersion    string
		TestDescription string
	}{
		{
			InputLog: logrus.New(),
			InputClient: &fake.RedisClient{
				InfoRes: &redis.StringCmd{},
			},
			InputConfig:     Config{},
			InputPretty:     false,
			InputVersion:    "0.0.1",
			TestDescription: "Should successfully perform a run without error",
		},
	}

	for _, test := range tests {
		g.Describe("Run()", func() {
			g.It(test.TestDescription, func() {
				Run(test.InputLog, test.InputClient, test.InputConfig, test.InputPretty, test.InputVersion)
				g.Assert(true).IsTrue()
			})
		})
	}
}

func TestInitRedisClient(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputConfig     Config
		TestDescription string
	}{
		{
			InputConfig: Config{
				RedisHost: "localhost",
				RedisPort: "6379",
				RedisPass: "",
				RedisDB:   "0",
			},
			TestDescription: "Should successfully create a redis client",
		},
	}

	for _, test := range tests {
		g.Describe("initRedisClient()", func() {
			g.It(test.TestDescription, func() {
				InitRedisClient(test.InputConfig)
				g.Assert(true).IsTrue()
			})
		})
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

func TestValidateConfig(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputLog        *logrus.Logger
		InputConfig     Config
		ExpectedConfig  Config
		TestDescription string
	}{
		{
			InputLog:    logrus.New(),
			InputConfig: Config{},
			ExpectedConfig: Config{
				RedisHost: "localhost",
				RedisPort: "6379",
			},
			TestDescription: "Should successfully set proper defaults when none are provided",
		},
		{
			InputLog: logrus.New(),
			InputConfig: Config{
				RedisHost: "10.0.0.1",
				RedisPort: "1234",
				RedisPass: "somepass",
				RedisDB:   "2",
			},
			ExpectedConfig: Config{
				RedisHost: "10.0.0.1",
				RedisPort: "1234",
				RedisPass: "somepass",
				RedisDB:   "2",
				DBID:      2,
			},
			TestDescription: "Should successfully set proper defaults when none are provided",
		},
	}

	for _, test := range tests {
		g.Describe("validateConfig()", func() {
			g.It(test.TestDescription, func() {
				ValidateConfig(test.InputLog, &test.InputConfig)
				g.Assert(test.InputConfig).Equal(test.ExpectedConfig)
			})
		})
	}
}

func TestReadStats(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputLog        *logrus.Logger
		InputClient     *fake.RedisClient
		InputConfig     Config
		ExpectedRes     string
		TestDescription string
	}{
		{
			InputLog: logrus.New(),
			InputClient: &fake.RedisClient{
				InfoRes: &redis.StringCmd{},
			},
			InputConfig:     Config{},
			ExpectedRes:     "",
			TestDescription: "Should successfully read stats from redis",
		},
	}

	for _, test := range tests {
		g.Describe("readStats()", func() {
			g.It(test.TestDescription, func() {
				res := readStats(test.InputLog, test.InputClient, test.InputConfig)
				g.Assert(res).Equal(test.ExpectedRes)
			})
		})
	}
}

func TestParseRawData(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputData       string
		ExpectedRes     map[string]string
		TestDescription string
	}{
		{
			InputData: "key1:value1\r\nkey2:value2\r\nkey3:value3\r\n",
			ExpectedRes: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
			TestDescription: "Should successfully parse and format output string into map",
		},
		{
			InputData:       "key1:value1\rkey2:value2\rkey3:value3\r",
			ExpectedRes:     map[string]string{},
			TestDescription: "Should return empty map if data is not formatted properly",
		},
	}

	for _, test := range tests {
		g.Describe("parseRawData()", func() {
			g.It(test.TestDescription, func() {
				res := parseRawData(test.InputData)
				g.Assert(res).Equal(test.ExpectedRes)
			})
		})
	}
}

func TestFormatMetric(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputLog        *logrus.Logger
		InputData       string
		ExpectedRes     []byte
		TestDescription string
	}{
		{
			InputLog:        logrus.New(),
			InputData:       "redis.redis_version:0.0.1\r\nredis.redis_git_sha1:00000000\r\nredis.redis_git_dirty:1\r\n",
			ExpectedRes:     []byte(`{"event_type":"RedisInfo","providor":"redis","redis.aof_current_rewrite_time_sec":0,"redis.aof_enabled":0,"redis.aof_last_bgrewrite_status":"","redis.aof_last_rewrite_time_sec":0,"redis.aof_last_write_status":"","redis.aof_rewrite_in_progress":0,"redis.aof_rewrite_scheduled":0,"redis.arch_bits":0,"redis.blocked_clients":0,"redis.client_biggest_input_buf":0,"redis.client_longest_output_list":0,"redis.cluster_enabled":0,"redis.config_file":"","redis.connected_clients":0,"redis.connected_slaves":0,"redis.evicted_keys":0,"redis.executable":"","redis.expired_keys":0,"redis.gcc_version":"","redis.hz":0,"redis.instantaneous_input_kbps":0,"redis.instantaneous_ops_per_sec":0,"redis.instantaneous_output_kbps":0,"redis.keyspace_hits":0,"redis.keyspace_misses":0,"redis.latest_fork_usec":0,"redis.loading":0,"redis.lru_clock":0,"redis.master_repl_offset":0,"redis.maxmemory":0,"redis.maxmemory_human":"","redis.maxmemory_policy":"","redis.mem_allocator":"","redis.mem_fragmentation_ratio":0,"redis.migrate_cached_sockets":0,"redis.multiplexing_api":"","redis.os":"","redis.process_id":0,"redis.pubsub_channels":0,"redis.pubsub_patterns":0,"redis.rdb_bgsave_in_progress":0,"redis.rdb_changes_since_last_save":0,"redis.rdb_current_bgsave_time_sec":0,"redis.rdb_last_bgsave_status":"","redis.rdb_last_bgsave_time_sec":0,"redis.rdb_last_save_time":0,"redis.redis_build_id":"","redis.redis_git_dirty":0,"redis.redis_git_sha1":"","redis.redis_mode":"","redis.redis_version":"","redis.rejected_connections":0,"redis.repl_backlog_active":0,"redis.repl_backlog_first_byte_offset":0,"redis.repl_backlog_histlen":0,"redis.repl_backlog_size":0,"redis.role":"","redis.run_id":"","redis.sync_full":0,"redis.sync_partial_err":0,"redis.sync_partial_ok":0,"redis.tcp_port":0,"redis.total_commands_processed":0,"redis.total_connections_received":0,"redis.total_net_input_bytes":0,"redis.total_net_output_bytes":0,"redis.total_system_memory":0,"redis.total_system_memory_human":"","redis.uptime_in_days":0,"redis.uptime_in_seconds":0,"redis.used_cpu_sys":0,"redis.used_cpu_sys_children":0,"redis.used_cpu_user":0,"redis.used_cpu_user_children":0,"redis.used_memory":0,"redis.used_memory_human":"","redis.used_memory_lua":0,"redis.used_memory_lua_human":"","redis.used_memory_peak":0,"redis.used_memory_peak_human":"","redis.used_memory_rss":0,"redis.used_memory_rss_human":""}`),
			TestDescription: "Should successfully parse and format output capable of being formatted into json",
		},
	}

	for _, test := range tests {
		g.Describe("formatMetric()", func() {
			g.It(test.TestDescription, func() {
				res := formatMetric(test.InputLog, test.InputData)
				js, err := json.Marshal(res)

				g.Assert(err).Equal(nil)
				g.Assert(string(js)).Equal(string(test.ExpectedRes))
			})
		})
	}
}

func TestToInt(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputLog        *logrus.Logger
		InputValue      string
		ExpectedRes     int
		TestDescription string
	}{
		{
			InputLog:        logrus.New(),
			InputValue:      "3",
			ExpectedRes:     3,
			TestDescription: "Should successfully parse to int when given valid int value",
		},
		{
			InputLog:        logrus.New(),
			InputValue:      "puppies",
			ExpectedRes:     0,
			TestDescription: "Should return 0 when given no int input",
		},
		{
			InputLog:        logrus.New(),
			InputValue:      "",
			ExpectedRes:     0,
			TestDescription: "Should return 0 when given blank input",
		},
	}

	for _, test := range tests {
		g.Describe("toInt()", func() {
			g.It(test.TestDescription, func() {
				res := toInt(test.InputLog, test.InputValue)
				g.Assert(res).Equal(test.ExpectedRes)
			})
		})
	}
}

func TestToFloat(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputLog        *logrus.Logger
		InputValue      string
		ExpectedRes     float64
		TestDescription string
	}{
		{
			InputLog:        logrus.New(),
			InputValue:      "3.14",
			ExpectedRes:     3.14,
			TestDescription: "Should successfully parse to float64 when given valid float value",
		},
		{
			InputLog:        logrus.New(),
			InputValue:      "puppies",
			ExpectedRes:     0,
			TestDescription: "Should return 0 when given no float input",
		},
		{
			InputLog:        logrus.New(),
			InputValue:      "",
			ExpectedRes:     0,
			TestDescription: "Should return 0 when given blank input",
		},
	}

	for _, test := range tests {
		g.Describe("toFloat()", func() {
			g.It(test.TestDescription, func() {
				res := toFloat(test.InputLog, test.InputValue)
				g.Assert(res).Equal(test.ExpectedRes)
			})
		})
	}
}
