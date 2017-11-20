package couchbase

import (
	"errors"
	"os"
	"reflect"
	"sync"
	"testing"

	fake "github.com/GannettDigital/paas-api-utils/utilsHTTP/fake"
	"github.com/Sirupsen/logrus"
	"github.com/franela/goblin"
)

var couchbaseFakeConfig CouchbaseConfig

func init() {
	couchbaseFakeConfig = CouchbaseConfig{
		CouchbaseHost:     "http://localhost",
		CouchbasePassword: "secure",
		CouchbasePort:     "15672",
		CouchbaseUser:     "admin",
	}
}

func TestValidateConfig(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputLog        *logrus.Logger
		InputConfig     CouchbaseConfig
		ExpectedErr     error
		TestDescription string
	}{
		{
			InputLog:        logrus.New(),
			InputConfig:     CouchbaseConfig{},
			ExpectedErr:     errors.New("Config Yaml is missing CouchbaseHost value. Please check the config to continue"),
			TestDescription: "Should Error when CouchbaseHost is not set",
		},
		{
			InputLog: logrus.New(),
			InputConfig: CouchbaseConfig{
				CouchbaseHost: "Derp",
			},
			ExpectedErr:     errors.New("Config Yaml is missing CouchbasePassword value. Please check the config to continue"),
			TestDescription: "Should Error when CouchbasePassword is not set",
		},
		{
			InputLog: logrus.New(),
			InputConfig: CouchbaseConfig{
				CouchbaseHost:     "Derp",
				CouchbasePassword: "Derp",
			},
			ExpectedErr:     errors.New("Config Yaml is missing CouchbasePort value. Please check the config to continue"),
			TestDescription: "Should Error when CouchbasPort is not set",
		},
		{
			InputLog: logrus.New(),
			InputConfig: CouchbaseConfig{
				CouchbaseHost:     "Derp",
				CouchbasePassword: "Derp",
				CouchbasePort:     "DerpPort",
			},
			ExpectedErr:     errors.New("Config Yaml is missing CouchbaseUser value. Please check the config to continue"),
			TestDescription: "Should Error when CouchbaseUser is not set",
		},
		{
			InputLog: logrus.New(),
			InputConfig: CouchbaseConfig{
				CouchbaseHost:     "Derp",
				CouchbasePassword: "Derp",
				CouchbasePort:     "DerpPort",
				CouchbaseUser:     "User",
			},
			ExpectedErr:     nil,
			TestDescription: "Should not error when all config values are set",
		},
	}

	for _, test := range tests {
		g.Describe("validateConfig()", func() {
			g.It(test.TestDescription, func() {
				err := validateConfig(test.InputLog, test.InputConfig)
				g.Assert(err).Equal(test.ExpectedErr)
			})
		})
	}
}

func TestGetCouchBucketsStats(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		HTTPRunner      fake.HTTPResult
		ExpectedResult  []map[string]interface{}
		TestDescription string
	}{
		{
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					fake.Result{
						Method: "GET",
						URI:    "/pools/default/buckets",
						Code:   200,
						Data:   []byte(`[{"name":"test1","uri":"/pools/default/buckets/test1","stats":{"uri":"/pools/default/buckets/test1/stats"}}]`),
					},
					fake.Result{
						Method: "GET",
						URI:    "/pools/default/buckets/test1/stats?zoom=minute",
						Code:   200,
						Data:   []byte(`{"op":{"samples":{"couch_total_disk_size":[368665193718,368665193718,368665193718],"couch_docs_fragmentation":[70,70,70],"couch_views_fragmentation":[0,0,0],"hit_ratio":[0,0,0],"ep_cache_miss_rate":[0,0,0],"ep_resident_items_rate":[57.32907676168598,57.32907676168598,57.32907676168598],"vb_avg_active_queue_age":[0,0,0],"vb_avg_replica_queue_age":[0,0,0],"vb_avg_pending_queue_age":[0,0,0],"vb_avg_total_queue_age":[0,0,0],"vb_active_resident_items_ratio":[57.32907676168598,57.32907676168598,57.32907676168598],"vb_replica_resident_items_ratio":[100,100,100],"vb_pending_resident_items_ratio":[100,100,100],"avg_disk_update_time":[0,0,0],"avg_disk_commit_time":[0,0.5,0.5],"avg_bg_wait_time":[0,0,0],"ep_dcp_views+indexes_count":[0,0,0],"ep_dcp_views+indexes_items_remaining":[0,0,0],"ep_dcp_views+indexes_producer_count":[0,0,0],"ep_dcp_views+indexes_total_backlog_size":[0,0,0],"ep_dcp_views+indexes_items_sent":[0,0,0],"ep_dcp_views+indexes_total_bytes":[0,0,0],"ep_dcp_views+indexes_backoff":[0,0,0],"bg_wait_count":[0,0,0],"bg_wait_total":[0,0,0],"bytes_read":[1050,711,771],"bytes_written":[178266,104371,118829],"cas_badval":[0,0,0],"cas_hits":[0,0,0],"cas_misses":[0,0,0],"cmd_get":[0,0,0],"cmd_set":[0,0,0],"couch_docs_actual_disk_size":[368665193718,368665193718,368665193718],"couch_docs_data_size":[105353370541,105353370541,105353370541],"couch_docs_disk_size":[355602721792,355602721792,355602721792],"couch_spatial_data_size":[0,0,0],"couch_spatial_disk_size":[0,0,0],"couch_spatial_ops":[0,0,0],"couch_views_actual_disk_size":[0,0,0],"couch_views_data_size":[0,0,0],"couch_views_disk_size":[0,0,0],"couch_views_ops":[0,0,0],"curr_connections":[115,115,115],"curr_items":[130735573,130735573,130735573],"curr_items_tot":[130735573,130735573,130735573],"decr_hits":[0,0,0],"decr_misses":[0,0,0],"delete_hits":[0,0,0],"delete_misses":[0,0,0],"disk_commit_count":[0,4,5],"disk_commit_total":[0,2000000,2500000],"disk_update_count":[0,0,0],"disk_update_total":[0,0,0],"disk_write_queue":[0,0,0],"ep_bg_fetched":[0,0,0],"ep_dcp_2i_backoff":[0,0,0],"ep_dcp_2i_count":[0,0,0],"ep_dcp_2i_items_remaining":[0,0,0],"ep_dcp_2i_items_sent":[0,0,0],"ep_dcp_2i_producer_count":[0,0,0],"ep_dcp_2i_total_backlog_size":[0,0,0],"ep_dcp_2i_total_bytes":[0,0,0],"ep_dcp_fts_backoff":[0,0,0],"ep_dcp_fts_count":[0,0,0],"ep_dcp_fts_items_remaining":[0,0,0],"ep_dcp_fts_items_sent":[0,0,0],"ep_dcp_fts_producer_count":[0,0,0],"ep_dcp_fts_total_backlog_size":[0,0,0],"ep_dcp_fts_total_bytes":[0,0,0],"ep_dcp_other_backoff":[0,0,0],"ep_dcp_other_count":[0,0,0],"ep_dcp_other_items_remaining":[0,0,0],"ep_dcp_other_items_sent":[0,0,0],"ep_dcp_other_producer_count":[0,0,0],"ep_dcp_other_total_backlog_size":[0,0,0],"ep_dcp_other_total_bytes":[0,0,0],"ep_dcp_replica_backoff":[0,0,0],"ep_dcp_replica_count":[0,0,0],"ep_dcp_replica_items_remaining":[0,0,0],"ep_dcp_replica_items_sent":[0,0,0],"ep_dcp_replica_producer_count":[0,0,0],"ep_dcp_replica_total_backlog_size":[0,0,0],"ep_dcp_replica_total_bytes":[0,0,0],"ep_dcp_views_backoff":[0,0,0],"ep_dcp_views_count":[0,0,0],"ep_dcp_views_items_remaining":[0,0,0],"ep_dcp_views_items_sent":[0,0,0],"ep_dcp_views_producer_count":[0,0,0],"ep_dcp_views_total_backlog_size":[0,0,0],"ep_dcp_views_total_bytes":[0,0,0],"ep_dcp_xdcr_backoff":[0,0,0],"ep_dcp_xdcr_count":[0,0,0],"ep_dcp_xdcr_items_remaining":[0,0,0],"ep_dcp_xdcr_items_sent":[0,0,0],"ep_dcp_xdcr_producer_count":[0,0,0],"ep_dcp_xdcr_total_backlog_size":[0,0,0],"ep_dcp_xdcr_total_bytes":[0,0,0],"ep_diskqueue_drain":[0,0,0],"ep_diskqueue_fill":[0,0,0],"ep_diskqueue_items":[0,0,0],"ep_flusher_todo":[0,0,0],"ep_item_commit_failed":[0,0,0],"ep_kv_size":[86469400604,86469400604,86469400604],"ep_max_size":[107374182400,107374182400,107374182400],"ep_mem_high_wat":[91268055040,91268055040,91268055040],"ep_mem_low_wat":[80530636800,80530636800,80530636800],"ep_meta_data_memory":[16882794956,16882794956,16882794956],"ep_num_non_resident":[55786076,55786076,55786076],"ep_num_ops_del_meta":[0,0,0],"ep_num_ops_del_ret_meta":[0,0,0],"ep_num_ops_get_meta":[0,0,0],"ep_num_ops_set_meta":[0,0,0],"ep_num_ops_set_ret_meta":[0,0,0],"ep_num_value_ejects":[0,0,0],"ep_oom_errors":[0,0,0],"ep_ops_create":[0,0,0],"ep_ops_update":[0,0,0],"ep_overhead":[809185792,809185792,809185792],"ep_queue_size":[0,0,0],"ep_tap_rebalance_count":[0,0,0],"ep_tap_rebalance_qlen":[0,0,0],"ep_tap_rebalance_queue_backfillremaining":[0,0,0],"ep_tap_rebalance_queue_backoff":[0,0,0],"ep_tap_rebalance_queue_drain":[0,0,0],"ep_tap_rebalance_queue_fill":[0,0,0],"ep_tap_rebalance_queue_itemondisk":[0,0,0],"ep_tap_rebalance_total_backlog_size":[0,0,0],"ep_tap_replica_count":[0,0,0],"ep_tap_replica_qlen":[0,0,0],"ep_tap_replica_queue_backfillremaining":[0,0,0],"ep_tap_replica_queue_backoff":[0,0,0],"ep_tap_replica_queue_drain":[0,0,0],"ep_tap_replica_queue_fill":[0,0,0],"ep_tap_replica_queue_itemondisk":[0,0,0],"ep_tap_replica_total_backlog_size":[0,0,0],"ep_tap_total_count":[0,0,0],"ep_tap_total_qlen":[0,0,0],"ep_tap_total_queue_backfillremaining":[0,0,0],"ep_tap_total_queue_backoff":[0,0,0],"ep_tap_total_queue_drain":[0,0,0],"ep_tap_total_queue_fill":[0,0,0],"ep_tap_total_queue_itemondisk":[0,0,0],"ep_tap_total_total_backlog_size":[0,0,0],"ep_tap_user_count":[0,0,0],"ep_tap_user_qlen":[0,0,0],"ep_tap_user_queue_backfillremaining":[0,0,0],"ep_tap_user_queue_backoff":[0,0,0],"ep_tap_user_queue_drain":[0,0,0],"ep_tap_user_queue_fill":[0,0,0],"ep_tap_user_queue_itemondisk":[0,0,0],"ep_tap_user_total_backlog_size":[0,0,0],"ep_tmp_oom_errors":[0,0,0],"ep_vb_total":[1024,1024,1024],"evictions":[0,0,0],"get_hits":[0,0,0],"get_misses":[0,0,0],"incr_hits":[0,0,0],"incr_misses":[0,0,0],"mem_used":[90771010096,90771010096,90771010096],"misses":[0,0,0],"ops":[0,0,0],"timestamp":[1488911853731,1488911854731,1488911855731],"vb_active_eject":[0,0,0],"vb_active_itm_memory":[79666497059,79666497059,79666497059],"vb_active_meta_data_memory":[16882794956,16882794956,16882794956],"vb_active_num":[1024,1024,1024],"vb_active_num_non_resident":[55786076,55786076,55786076],"vb_active_ops_create":[0,0,0],"vb_active_ops_update":[0,0,0],"vb_active_queue_age":[0,0,0],"vb_active_queue_drain":[0,0,0],"vb_active_queue_fill":[0,0,0],"vb_active_queue_size":[0,0,0],"vb_pending_curr_items":[0,0,0],"vb_pending_eject":[0,0,0],"vb_pending_itm_memory":[0,0,0],"vb_pending_meta_data_memory":[0,0,0],"vb_pending_num":[0,0,0],"vb_pending_num_non_resident":[0,0,0],"vb_pending_ops_create":[0,0,0],"vb_pending_ops_update":[0,0,0],"vb_pending_queue_age":[0,0,0],"vb_pending_queue_drain":[0,0,0],"vb_pending_queue_fill":[0,0,0],"vb_pending_queue_size":[0,0,0],"vb_replica_curr_items":[0,0,0],"vb_replica_eject":[0,0,0],"vb_replica_itm_memory":[0,0,0],"vb_replica_meta_data_memory":[0,0,0],"vb_replica_num":[0,0,0],"vb_replica_num_non_resident":[0,0,0],"vb_replica_ops_create":[0,0,0],"vb_replica_ops_update":[0,0,0],"vb_replica_queue_age":[0,0,0],"vb_replica_queue_drain":[0,0,0],"vb_replica_queue_fill":[0,0,0],"vb_replica_queue_size":[0,0,0],"vb_total_queue_age":[0,0,0],"xdc_ops":[0,0,0],"cpu_idle_ms":[39260,39140,38760],"cpu_local_ms":[39830,39830,39800],"cpu_utilization_rate":[2.512562814070352,2.893081761006289,4.020100502512562],"hibernated_requests":[20,20,20],"hibernated_waked":[0,0,0],"mem_actual_free":[144403140608,144410914816,143918616576],"mem_actual_used":[177665318912,177657544704,177674334208],"mem_free":[144403140608,144410914816,144394125312],"mem_total":[322068459520,322068459520,322068459520],"mem_used_sys":[313455263744,313447501824,313464295424],"rest_requests":[21,35,6],"swap_total":[0,0,0],"swap_used":[0,0,0]}}}`),
					},
				},
			},
			ExpectedResult:  []map[string]interface{}{},
			TestDescription: "Successfully GET Couchbase  bucket stats",
		},
	}

	for _, test := range tests {
		g.Describe("TestGetCouchBucketsStats()", func() {
			g.It(test.TestDescription, func() {
				runner = &test.HTTPRunner
				couchBucketResponses, getCouchBucketStatsError := getCouchBucketsStats(logrus.New(), couchbaseFakeConfig)
				g.Assert(getCouchBucketStatsError).Equal(nil)
				g.Assert(len(couchBucketResponses)).Equal(2)
			})
		})
	}
}

func TestGetCouchClusterStats(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		HTTPRunner        fake.HTTPResult
		ExpectedScalrName string
		ExpectedResult    []map[string]interface{}
		TestDescription   string
	}{
		{
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					fake.Result{
						Method: "GET",
						URI:    "/pools/default",
						Code:   200,
						Data:   []byte(`{"indexStatusURI":"indexStatus?v=74148774","storageTotals":{"ram":{"total":1,"quotaTotal":22,"quotaUsed":333,"used":4444,"usedByData":55555,"quotaUsedPerNode":666666,"quotaTotalPerNode":7777777},"hdd":{"total":1,"quotaTotal":22,"used":333,"usedByData":4444,"free":55555}},"nodes":[{"systemStats":{"cpu_utilization_rate":0.7731958762886598,"swap_total":4192202752,"swap_used":0,"mem_total":2095890432,"mem_free":1118621696},"interestingStats":{"cmd_get":0,"couch_docs_actual_disk_size":23087235,"couch_docs_data_size":14033920,"couch_spatial_data_size":0,"couch_spatial_disk_size":0,"couch_views_actual_disk_size":2792937,"couch_views_data_size":2792937,"curr_items":7303,"curr_items_tot":7303,"ep_bg_fetched":0,"get_hits":0,"mem_used":104462392,"ops":0,"vb_replica_curr_items":0},"uptime":"4963","memoryTotal":2095890432,"memoryFree":1118621696,"mcdMemoryReserved":1599,"mcdMemoryAllocated":1599,"couchApiBase":"http://172.17.0.2:8092/","couchApiBaseHTTPS":"https://172.17.0.2:18092/","clusterMembership":"active","recoveryType":"none","status":"healthy","otpNode":"ns_1@127.0.0.1","thisNode":true,"hostname":"172.17.0.2:8091","clusterCompatibility":262149,"version":"4.5.1-2844-enterprise","os":"x86_64-unknown-linux-gnu","ports":{"sslProxy":11214,"httpsMgmt":18091,"httpsCAPI":18092,"proxy":11211,"direct":11210},"services":["index","kv","n1ql"]}]}`),
					},
					fake.Result{
						Method: "GET",
						URI:    "/indexStatus?v=74148774",
						Code:   200,
						Data:   []byte(`{"indexes":[{"storageMode":"forestdb","hosts":["10.84.96.220:8091"],"progress":100,"definition":"CREATE PRIMARY INDEX #primary ON deployments","status":"Ready","bucket":"deployments","index":"#primary","id":8893382884342514204},{"storageMode":"forestdb","hosts":["10.84.96.220:8091"],"progress":100,"definition":"CREATE PRIMARY INDEX #primary ON dns-manager","status":"Ready","bucket":"dns-manager","index":"#primary","id":7168709155480627845}]}`),
					},
				},
			},
			ExpectedScalrName: "sometestname",
			ExpectedResult:    []map[string]interface{}{},
			TestDescription:   "Successfully GET Couchbase Cluster stats",
		},
	}

	for _, test := range tests {
		origValue := os.Getenv("CB_CLUSTER_NAME")
		os.Setenv("CB_CLUSTER_NAME", test.ExpectedScalrName)

		g.Describe("TestGetCouchBucketsStats()", func() {
			g.It(test.TestDescription, func() {
				runner = &test.HTTPRunner
				couchBucketResponses, getCouchBucketStatsError := getCouchClusterStats(logrus.New(), couchbaseFakeConfig)
				g.Assert(getCouchBucketStatsError).Equal(nil)
				g.Assert(couchBucketResponses[3]["couchbase.scalr.clustername"]).Equal(test.ExpectedScalrName)
				g.Assert(couchBucketResponses[3]["couchbase.cluster.hdd.free"]).Equal(int64(55555))
				g.Assert(couchBucketResponses[3]["event_type"]).Equal(EVENT_TYPE)
				g.Assert(couchBucketResponses[3]["provider"]).Equal(PROVIDER)
				g.Assert(couchBucketResponses[0]["couchbase.cluster.by_node.status"]).Equal("healthy")
				g.Assert(couchBucketResponses[0]["couchbase.cluster.by_node.hostname"]).Equal("172.17.0.2:8091")
				g.Assert(couchBucketResponses[0]["couchbase.cluster.by_node.cluster_membership"]).Equal("active")
			})
		})

		os.Setenv("CB_CLUSTER_NAME", origValue)
	}
}

func TestGetCouchReplicationStats(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		HTTPRunner        fake.HTTPResult
		ExpectedScalrName string
		ExpectedResult    []map[string]interface{}
		TestDescription   string
	}{
		{
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					fake.Result{
						Method: "GET",
						URI:    "/pools/default/remoteClusters",
						Code:   200,
						Data:   []byte(`[{"deleted":false,"hostname":"172.17.0.2:8091","name":"paas-api-west","uri":"/pools/default/remoteClusters/paas-api-west","username":"paas-api","uuid":"derp","validateURI":"/pools/default/remoteClusters/paas-api-west?just_validate=1"}]`),
					},
				},
			},
			ExpectedScalrName: "sometestname",
			ExpectedResult:    []map[string]interface{}{},
			TestDescription:   "Successfully GET Couchbase Cluster stats",
		},
	}

	for _, test := range tests {
		origValue := os.Getenv("CB_CLUSTER_NAME")
		os.Setenv("CB_CLUSTER_NAME", test.ExpectedScalrName)

		g.Describe("TestGetCouchReplicationStats()", func() {
			g.It(test.TestDescription, func() {
				runner = &test.HTTPRunner
				couchReplicationResponses, getCouchReplicationStatsError := getCouchReplicationStats(logrus.New(), couchbaseFakeConfig)
				g.Assert(getCouchReplicationStatsError).Equal(nil)
				g.Assert(couchReplicationResponses[0]["couchbase.replication.hostname"]).Equal("172.17.0.2:8091")
				g.Assert(couchReplicationResponses[0]["couchbase.replication.uuid"]).Equal("derp")
			})
		})

		os.Setenv("CB_CLUSTER_NAME", origValue)
	}
}

func TestGetAllBucketsInfo(t *testing.T) {
	g := goblin.Goblin(t)

	type AllBucketsInfoTestResults struct {
		ErrorShouldBeNil      bool
		ClusterInfoShouldHave int
	}
	var tests = []struct {
		HTTPRunner      fake.HTTPResult
		TestResults     AllBucketsInfoTestResults
		TestDescription string
	}{
		{
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					fake.Result{
						Method: "GET",
						URI:    "/pools/default/buckets",
						Code:   200,
						Data:   []byte("[{\"name\":\"test1\",\"uri\":\"/pools/default/buckets/test1\",\"stats\":{\"uri\":\"/pools/default/buckets/test1/stats\"}},{\"name\":\"test2\",\"uri\":\"/pools/default/buckets/test2\",\"stats\":{\"uri\":\"/pools/default/buckets/test2/stats\"}}]"),
					},
				},
			},
			TestResults: AllBucketsInfoTestResults{
				ErrorShouldBeNil:      true,
				ClusterInfoShouldHave: 2,
			},
			TestDescription: "Successfully GET List of buckets",
		},
		{
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					fake.Result{
						Method: "GET",
						URI:    "/pools/default/buckets",
						Code:   500,
						Err:    errors.New("Internal Error"),
						Data:   []byte("[{\"name\":\"test1\",\"uri\":\"/pools/default/buckets/test1\",\"stats\":{\"uri\":\"/pools/default/buckets/test1/stats\"}},{\"name\":\"test2\",\"uri\":\"/pools/default/buckets/test2\",\"stats\":{\"uri\":\"/pools/default/buckets/test2/stats\"}}]"),
					},
				},
			},
			TestResults: AllBucketsInfoTestResults{
				ErrorShouldBeNil:      false,
				ClusterInfoShouldHave: 0,
			},
			TestDescription: "Fail GET List of buckets",
		},
	}

	for _, test := range tests {
		g.Describe("getAllBucketsInfo()", func() {
			g.It(test.TestDescription, func() {
				runner = &test.HTTPRunner
				result, err := getAllBucketsInfo(logrus.New(), couchbaseFakeConfig)
				g.Assert(err == nil).Equal(test.TestResults.ErrorShouldBeNil)
				g.Assert(reflect.DeepEqual(len(result), test.TestResults.ClusterInfoShouldHave)).Equal(true)
			})
		})
	}
}

func TestAvgInt64Sample(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		TestDescription string
		TestData        []int64
		ExpectedResult  float32
	}{
		{
			TestDescription: "Successfully Avg int64 [355602721792, 355602721792, 355602721792]",
			TestData:        []int64{355602721792, 355602721792, 355602721792},
			ExpectedResult:  3.5560273e+11,
		},
		{
			TestDescription: "Successfully Avg int64 []",
			TestData:        []int64{},
			ExpectedResult:  0,
		},
	}

	for _, test := range tests {
		g.Describe("avgInt64Sample()", func() {
			g.It(test.TestDescription, func() {
				result := avgInt64Sample(test.TestData)
				g.Assert(reflect.DeepEqual(result, test.ExpectedResult)).Equal(true)
			})
		})
	}
}

func TestAvgFloat32Sample(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		TestDescription string
		TestData        []float32
		ExpectedResult  float32
	}{
		{
			TestDescription: "Successfully Avg float32 [100.00, 250.33, 98.12]",
			TestData:        []float32{100.00, 250.33, 98.12},
			ExpectedResult:  149.48334,
		},
		{
			TestDescription: "Successfully Avg int []",
			TestData:        []float32{},
			ExpectedResult:  0,
		},
	}

	for _, test := range tests {
		g.Describe("avgFloat32Sample()", func() {
			g.It(test.TestDescription, func() {
				result := avgFloat32Sample(test.TestData)
				g.Assert(reflect.DeepEqual(result, test.ExpectedResult)).Equal(true)
			})
		})
	}
}

func TestFormatBucketInfoStructToMap(t *testing.T) {
	g := goblin.Goblin(t)

	var couchbucketStats CouchbaseBucketStats
	couchbucketStats.OP.Samples.GetHits = []float32{2, 2, 2}
	var tests = []struct {
		TestDescription string
		TestData        CompleteBucketInfo
		ExpectedResult  map[string]interface{}
	}{
		{
			TestDescription: "Successfully Convert Complete Bucket info to Map Interface",
			TestData: CompleteBucketInfo{
				bucketInfo: CouchbaseBucketStatsURI{
					Name: "TestName",
				},
				bucketStats: couchbucketStats,
			},
			ExpectedResult: map[string]interface{}{
				"couchbase.by_bucket.name":     "TestName",
				"couchbase.by_bucket.get_hits": float32(2),
			},
		},
	}

	for _, test := range tests {
		g.Describe("formatBucketInfoStructToMap()", func() {
			g.It(test.TestDescription, func() {
				result := formatBucketInfoStatsStructToMap(test.TestData)
				g.Assert(result["couchbase.by_bucket.name"] == test.ExpectedResult["couchbase.by_bucket.name"]).Equal(true)
				g.Assert(result["couchbase.by_bucket.get_hits"] == test.ExpectedResult["couchbase.by_bucket.get_hits"]).Equal(true)
			})
		})
	}
}

func TestFormatBucketInfoEPStatsStructToMap(t *testing.T) {
	g := goblin.Goblin(t)

	var couchbucketStats CouchbaseBucketStats
	couchbucketStats.OP.Samples.EPBGFetched = []float32{2, 2, 2}
	var tests = []struct {
		TestDescription string
		TestData        CompleteBucketInfo
		ExpectedResult  map[string]interface{}
	}{
		{
			TestDescription: "Successfully Convert Complete Bucket info to Map Interface",
			TestData: CompleteBucketInfo{
				bucketInfo: CouchbaseBucketStatsURI{
					Name: "TestName",
				},
				bucketStats: couchbucketStats,
			},
			ExpectedResult: map[string]interface{}{
				"couchbase.by_bucket.name":          "TestName",
				"couchbase.by_bucket.ep_bg_fetched": float32(2),
			},
		},
	}

	for _, test := range tests {
		g.Describe("formatBucketInfoEPStatsStructToMap()", func() {
			g.It(test.TestDescription, func() {
				result := formatBucketInfoEPStatsStructToMap(test.TestData)
				g.Assert(result["couchbase.by_bucket.name"] == test.ExpectedResult["couchbase.by_bucket.name"]).Equal(true)
				g.Assert(result["couchbase.by_bucket.get_hits"] == test.ExpectedResult["couchbase.by_bucket.get_hits"]).Equal(true)
			})
		})
	}
}

func TestGetCouchRemoteReplicationStats(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		TestDescription string
		HTTPRunner      fake.HTTPResult
		InputLog        *logrus.Logger
		InputConfig     CouchbaseConfig
		InputBuckets    []string
		InputUUIDs      []string
		InputEndpoints  []string
		ExpectedData    []MetricData
		ExpectedErr     error
	}{
		{
			TestDescription: "Should successfully process a remote stats request",
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					{
						Method: "GET",
						URI:    "/pools/default/buckets/deployments/stats/replications%2Fsomeuuid%2Fdeployments%2Fdeployments%2fsome_stats",
						Code:   200,
						Err:    nil,
						Data:   []byte(`{"samplesCount": 60,"isPersistent": true,"lastTStamp": 1510956143435,"interval": 1000,"timestamp": [1510956085435],"nodeStats": {"10.84.87.226:8091": [0]}}`),
					},
					{
						Method: "GET",
						URI:    "/pools/default/buckets/stuff/stats/replications%2Fsomeuuid%2Fstuff%2Fstuff%2fsome_stats",
						Code:   200,
						Err:    nil,
						Data:   []byte(`{"samplesCount": 60,"isPersistent": true,"lastTStamp": 1510956143435,"interval": 1000,"timestamp": [1510956085435],"nodeStats": {"10.84.87.226:8091": [0]}}`),
					},
				},
			},
			InputLog: logrus.New(),
			InputConfig: CouchbaseConfig{
				CouchbaseHost:     "http://derp.com",
				CouchbasePort:     "8091",
				CouchbaseUser:     "derp",
				CouchbasePassword: "derp",
			},
			InputBuckets: []string{
				"deployments",
				"stuff",
			},
			InputUUIDs:     []string{"someuuid"},
			InputEndpoints: []string{"some_stats"},
			ExpectedData: []MetricData{
				{
					"event_type": "CouchbaseReplicationSample",
					"provider":   "couchbase",
					"couchbase.replication.some_stats.samplescount": 60,
					"couchbase.replication.some_stats.ispersistent": true,
					"couchbase.replication.some_stats.lasttstamp":   1510956143435,
					"couchbase.replication.some_stats.interval":     1000,
					"couchbase.replication.some_stats.timestamp":    []int64{1510956085435},
					"couchbase.replication.some_stats.nodestats": map[string][]int64{
						"10.84.87.226:8091": []int64{0},
					},
				},
				{
					"event_type": "CouchbaseReplicationSample",
					"provider":   "couchbase",
					"couchbase.replication.some_stats.samplescount": 60,
					"couchbase.replication.some_stats.ispersistent": true,
					"couchbase.replication.some_stats.lasttstamp":   1510956143435,
					"couchbase.replication.some_stats.interval":     1000,
					"couchbase.replication.some_stats.timestamp":    []int64{1510956085435},
					"couchbase.replication.some_stats.nodestats": map[string][]int64{
						"10.84.87.226:8091": []int64{0},
					},
				},
			},
		},
	}

	for _, test := range tests {
		g.Describe("getCouchRemoteReplicationStats()", func() {
			g.It(test.TestDescription, func() {
				runner = test.HTTPRunner
				bucketList = test.InputBuckets
				remoteUUIDList = test.InputUUIDs
				remoteStatEndpoints = test.InputEndpoints
				data, err := getCouchRemoteReplicationStats(test.InputLog, test.InputConfig)
				g.Assert(len(data)).Equal(len(test.ExpectedData))
				g.Assert(err).Equal(test.ExpectedErr)
			})
		})
	}
}

func TestProcessRemoteReplicationStats(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		TestDescription string
		HTTPRunner      fake.HTTPResult
		InputLog        *logrus.Logger
		InputConfig     CouchbaseConfig
		InputBucket     string
		InputUUID       string
		InputEndpoint   string
		ExpectedResults remoteMeticChanResp
	}{
		{
			TestDescription: "Should successfully process a remote stats request",
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					{
						Method: "GET",
						URI:    "/pools/default/buckets/deployments/stats/replications%2Fsomeuuid%2Fdeployments%2Fdeployments%2fsome_stats",
						Code:   200,
						Err:    nil,
						Data:   []byte(`{"samplesCount": 60,"isPersistent": true,"lastTStamp": 1510956143435,"interval": 1000,"timestamp": [1510956085435],"nodeStats": {"10.84.87.226:8091": [0]}}`),
					},
				},
			},
			InputLog: logrus.New(),
			InputConfig: CouchbaseConfig{
				CouchbaseHost:     "http://derp.com",
				CouchbasePort:     "8091",
				CouchbaseUser:     "derp",
				CouchbasePassword: "derp",
			},
			InputBucket:   "deployments",
			InputUUID:     "someuuid",
			InputEndpoint: "some_stats",
			ExpectedResults: remoteMeticChanResp{
				Data: MetricData{
					"event_type": "CouchbaseReplicationSample",
					"provider":   "couchbase",
					"couchbase.replication.some_stats.samplescount": 60,
					"couchbase.replication.some_stats.ispersistent": true,
					"couchbase.replication.some_stats.lasttstamp":   1510956143435,
					"couchbase.replication.some_stats.interval":     1000,
					"couchbase.replication.some_stats.timestamp":    []int64{1510956085435},
					"couchbase.replication.some_stats.nodestats": map[string][]int64{
						"10.84.87.226:8091": []int64{0},
					},
				},
				Err: nil,
			},
		},
		{
			TestDescription: "Should return an error when one is encountered calling the endpoint",
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					{
						Method: "GET",
						URI:    "/pools/default/buckets/deployments/stats/replications%2Fsomeuuid%2Fdeployments%2Fdeployments%2fsome_stats",
						Code:   500,
						Err:    errors.New("some error"),
						Data:   []byte(``),
					},
				},
			},
			InputLog: logrus.New(),
			InputConfig: CouchbaseConfig{
				CouchbaseHost:     "http://derp.com",
				CouchbasePort:     "8091",
				CouchbaseUser:     "derp",
				CouchbasePassword: "derp",
			},
			InputBucket:   "deployments",
			InputUUID:     "someuuid",
			InputEndpoint: "some_stats",
			ExpectedResults: remoteMeticChanResp{
				Data: MetricData{},
				Err:  errors.New("some error"),
			},
		},
	}

	for _, test := range tests {
		g.Describe("processRemoteReplicationStats()", func() {
			runner = test.HTTPRunner
			g.It(test.TestDescription, func() {
				wg := &sync.WaitGroup{}
				inputChan := make(chan remoteMeticChanResp)
				wg.Add(1)
				go processRemoteReplicationStats(test.InputLog, test.InputConfig, wg, inputChan, test.InputBucket, test.InputUUID, test.InputEndpoint)
				go func() {
					wg.Wait()
					close(inputChan)
				}()

				for data := range inputChan {
					g.Assert(data.Err).Equal(test.ExpectedResults.Err)
				}
			})
		})
	}
}
