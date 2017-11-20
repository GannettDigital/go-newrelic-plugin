package couchbase

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/GannettDigital/go-newrelic-plugin/helpers"
	"github.com/GannettDigital/paas-api-utils/utilsHTTP"
	"github.com/Sirupsen/logrus"
)

var runner utilsHTTP.HTTPRunner
var bucketList, remoteUUIDList, remoteStatEndpoints []string

const EVENT_TYPE string = "DatastoreSample"
const NAME string = "couchbase"
const PROVIDER string = "couchbase"
const PROTOCOL_VERSION string = "1"

//CouchbaseConfig is the keeper of the config
type CouchbaseConfig struct {
	CouchbaseUser     string
	CouchbasePassword string
	CouchbasePort     string
	CouchbaseHost     string
}

type CouchbaseBucketStats struct {
	OP struct {
		Samples struct {
			AVGBGWaitTime               []float32 `json:"avg_bg_wait_time"`                //seconds
			AVGDiskCommitTime           []float32 `json:"avg_disk_commit_time"`            //seconds
			BytesRead                   []float32 `json:"bytes_read"`                      //bytes
			BytesWritten                []float32 `json:"bytes_written"`                   //bytes
			CasHits                     []float32 `json:"cas_hits"`                        //hits
			CasMisses                   []float32 `json:"cas_misses"`                      //misses
			CMDGet                      []float32 `json:"cmd_get"`                         //gets
			CMDSet                      []float32 `json:"cmd_set"`                         //sets
			CouchDocsActualDiskSize     []int64   `json:"couch_docs_actual_disk_size"`     //bytes
			CouchDocsDataSize           []int64   `json:"couch_docs_data_size"`            //bytes
			CouchDocsDiskSize           []int64   `json:"couch_docs_disk_size"`            //bytes
			CouchDocsFragmentation      []float32 `json:"couch_docs_fragmentation"`        //percent
			CouchTotalDiskSize          []int64   `json:"couch_total_disk_size"`           //bytes
			CouchViewsFragmentation     []float32 `json:"couch_views_fragmentation"`       //percent
			CouchViewsOps               []float32 `json:"couch_views_ops"`                 //operations
			CPUIdleTime                 []float32 `json:"cpu_idle_ms"`                     //milliseconds
			CPUUtilizationRate          []float32 `json:"cpu_utilization_rate"`            //percent
			CurrConnections             []float32 `json:"curr_connections"`                //connections
			CurrItems                   []float32 `json:"curr_items"`                      //items
			CurrItemsTotal              []float32 `json:"curr_items_tot"`                  //items
			DecrHits                    []float32 `json:"decr_hits"`                       //hits
			DecrMisses                  []float32 `json:"decr_misses"`                     //misses
			DeleteHits                  []float32 `json:"delete_hits"`                     //hits
			DeleteMisses                []float32 `json:"delete_misses"`                   //misses
			DiskCommitCount             []float32 `json:"disk_commit_count"`               //operations
			DiskUpdateCount             []float32 `json:"disk_update_count"`               //operations
			DiskWriteQueue              []float32 `json:"disk_write_queue"`                //operations
			Evictions                   []float32 `json:"evictions"`                       //evictions
			GetHits                     []float32 `json:"get_hits"`                        //hits
			GetMisses                   []float32 `json:"get_misses"`                      //misses
			HitRatio                    []float32 `json:"hit_ratio"`                       //percent
			IncrHits                    []float32 `json:"incr_hits"`                       //hits
			MemFree                     []int64   `json:"mem_free"`                        //bytes
			MemActuallFree              []int64   `json:"mem_actual_free"`                 //bytes
			MemTotal                    []int64   `json:"mem_total"`                       //bytes
			MemUsed                     []int64   `json:"mem_used"`                        //bytes
			MemActuallUsed              []int64   `json:"mem_actual_used"`                 //bytes
			Misses                      []float32 `json:"misses"`                          //misses
			Ops                         []float32 `json:"ops"`                             //operations
			VBActiveItmMemory           []float32 `json:"vb_active_itm_memory"`            //bytes
			VBActiveMetaDataMemory      []float32 `json:"vb_active_meta_data_memory"`      //bytes
			VBActiveNums                []float32 `json:"vb_active_num"`                   //items
			VBActiveQueueDrain          []float32 `json:"vb_active_queue_drain"`           //items
			VBActiveQueueSize           []float32 `json:"vb_active_queue_size"`            //items
			VBActiveResidentItemsRatio  []float32 `json:"vb_active_resident_items_ratio"`  //items
			VBActiveNumNonResident      []float32 `json:"vb_active_num_non_resident"`      //items
			VBAvgTotalQueueAge          []float32 `json:"vb_avg_total_queue_age"`          //Seconds
			VBPendingOpsCreate          []float32 `json:"vb_pending_ops_create"`           //operations
			VBPendingQueueFill          []float32 `json:"vb_pending_queue_fill"`           //items
			VBReplicaCurrItems          []float32 `json:"vb_replica_curr_items"`           //items
			VBReplicaItmMemory          []float32 `json:"vb_replica_itm_memory"`           //bytes
			VBReplicaMetaDataMemory     []float32 `json:"vb_replica_meta_data_memory"`     //bytes
			VBReplicaResidentItemsRatio []float32 `json:"vb_replica_resident_items_ratio"` //itmes
			VBReplicaNum                []float32 `json:"vb_replica_num"`                  //items
			VBReplicaQueueSize          []float32 `json:"vb_replica_queue_size"`           //items
			XDCOPS                      []float32 `json:"xdc_ops"`                         //operations
			EPBGFetched                 []float32 `json:"ep_bg_fetched"`                   //fetchs/second
			EPCacheMissRate             []float32 `json:"ep_cache_miss_rate"`              //misses
			EPDcp2iItemsRemaining       []int64   `json:"ep_dcp_2i_items_remaining"`       //items
			EPDcpFtsItemsRemaining      []int64   `json:"ep_dcp_fts_items_remaining"`      //items
			EPDcpOtherItemsRemaining    []int64   `json:"ep_dcp_other_items_remaining"`    //items
			EPDcpReplicaItemsRemaining  []int64   `json:"ep_dcp_replica_items_remaining"`  //items
			EPDcpReplicaItemsSent       []float32 `json:"ep_dcp_replica_items_sent"`       //items
			EPDcpReplicaTotalBytes      []float32 `json:"ep_dcp_replica_total_bytes"`      //bytes
			EPDcpViewItemsRemaining     []int64   `json:"ep_dcp_views_items_remaining"`    //items
			EPDcpXDCRItemsRemaining     []int64   `json:"ep_dcp_xdcr_items_remaining"`     //items
			EPDcpXDCRItemsSent          []float32 `json:"ep_dcp_xdcr_items_sent"`          //items
			EPDcpXDCRTotalBytes         []float32 `json:"ep_dcp_xdcr_total_bytes"`         //bytes
			EPDiskQueueItems            []float32 `json:"ep_diskqueue_items"`              //items
			EPDiskQueueDrain            []float32 `json:"ep_diskqueue_drain"`              //items
			EPDiskQueueFill             []float32 `json:"ep_diskqueue_fill"`               //items
			EPFlusherTodo               []float32 `json:"ep_flusher_todo"`                 //items
			EpItemCommitFailed          []float32 `json:"ep_item_commit_failed"`           //errors
			EPKVSize                    []int64   `json:"ep_kv_size"`                      //bytes
			EPMaxSize                   []int64   `json:"ep_max_size"`                     //bytes
			EPMemHighWater              []int64   `json:"ep_mem_high_wat"`                 //bytes
			EPMemLowWater               []int64   `json:"ep_mem_low_wat"`                  //bytes
			EPMetaDataMemory            []float32 `json:"ep_meta_data_memory"`             //bytes
			EPNumNonResident            []float32 `json:"ep_num_non_resident"`             //Items
			EPNumOpsGetMeta             []float32 `json:"ep_num_ops_get_meta"`             //operations
			EPNumOpsSetMeta             []float32 `json:"ep_num_ops_set_meta"`             //operations
			EPNumValueEjects            []float32 `json:"ep_num_value_ejects"`             //Items
			EPOOMErrors                 []float32 `json:"ep_oom_errors"`                   //errors
			EPOPSCreate                 []float32 `json:"ep_ops_create"`                   //operations
			EPOPSUpdate                 []float32 `json:"ep_ops_update"`                   //operations
			EPOverhead                  []int64   `json:"ep_overhead"`                     //bytes
			EPQueueSize                 []float32 `json:"ep_queue_size"`                   //items
			EPResidentItemsRate         []float32 `json:"ep_resident_items_rate"`          //items
			EPTapReplicaQueueDrain      []float32 `json:"ep_tap_replica_queue_drain"`      //items
			EPTapTotalQueueDrain        []float32 `json:"ep_tap_total_queue_drain"`        //items
			EPTapTotalQueueFill         []float32 `json:"ep_tap_total_queue_fill"`         //items
			EPTapTotalTotalBacklogSize  []float32 `json:"ep_tap_total_total_backlog_size"` //items
			EPTMPOOMErrors              []float32 `json:"ep_tmp_oom_errors"`               //errors
		} `json:"samples"`
	} `json:"op"`
}

type CouchbaseIndex struct {
	ID         uint64 `json:"id"`
	Bucket     string `json:"bucket"`
	Index      string `json:"index"`
	Status     string `json:"status"`
	Definition string `json:"definition"`
	Progress   int8   `json:"progress"`
}

type CouchbaseBucketStatsURI struct {
	Name        string `json:"name"`
	URI         string `json:"uri"`
	StatsObject struct {
		URI string `json:"uri"`
	} `json:"stats"`
}

type CouchbaseClusterInfo struct {
	Name           string `json:"name"`
	IndexStatusURI string `json:"indexStatusURI"`
	StorageTotals  struct {
		HDD struct {
			HDDFree       int64 `json:"free"`
			HDDTotal      int64 `json:"total"`
			HDDUsed       int64 `json:"used"`
			HDDUsedByData int64 `json:"usedByData"`
			HDDQuotaTotal int64 `json:"quotaTotal"`
		} `json:"hdd"`
		RAM struct {
			RAMUsed       int64 `json:"used"`
			RAMTotal      int64 `json:"total"`
			RAMUsedByData int64 `json:"usedByData"`
			RAMQuotaTotal int64 `json:"quotaTotal"`
		} `json:"ram"`
	} `json:"storageTotals"`
	Nodes []struct {
		HostName          string `json:"hostname"`
		ClusterMembership string `json:"clusterMembership"`
		Status            string `json:"status"`
	} `json:"nodes"`
}

type CompleteBucketInfo struct {
	bucketInfo  CouchbaseBucketStatsURI
	bucketStats CouchbaseBucketStats
}

type CouchbaseRemoteReplicationStats struct {
	SamplesCount int                `json:"samplesCount"`
	IsPersistent bool               `json:"isPersistent"`
	LastTStamp   int64              `json:"lastTStamp"`
	Internal     int                `json:"interval"`
	Timestamp    []int64            `json:"timestamp"`
	NodStats     map[string][]int64 `json:"nodeStats"`
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

func validateConfig(log *logrus.Logger, config CouchbaseConfig) error {
	if config.CouchbaseHost == "" {
		return errors.New("Config Yaml is missing CouchbaseHost value. Please check the config to continue")
	}
	if config.CouchbasePassword == "" {
		return errors.New("Config Yaml is missing CouchbasePassword value. Please check the config to continue")
	}
	if config.CouchbasePort == "" {
		return errors.New("Config Yaml is missing CouchbasePort value. Please check the config to continue")
	}
	if config.CouchbaseUser == "" {
		return errors.New("Config Yaml is missing CouchbaseUser value. Please check the config to continue")
	}
	return nil
}

func fatalIfErr(log *logrus.Logger, err error) {
	if err != nil {
		log.WithError(err).Fatal("can't continue")
	}
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

func init() {
	runner = &utilsHTTP.HTTPRunnerImpl{}
	bucketList = []string{}
	remoteUUIDList = []string{}
	remoteStatEndpoints = []string{
		"changes_left",
		"rate_replicated",
		"docs_written",
		"docs_checked",
		"docs_rep_queue",
		"num_checkpoints",
		"num_failedckpts",
		"bandwidth_usage",
	}
}

func Run(log *logrus.Logger, prettyPrint bool, version string) {

	// Initialize the output structure
	var data = PluginData{
		Name:            NAME,
		ProtocolVersion: PROTOCOL_VERSION,
		PluginVersion:   version,
		Inventory:       make(map[string]InventoryData),
		Metrics:         make([]MetricData, 0),
		Events:          make([]EventData, 0),
	}

	var config = CouchbaseConfig{
		CouchbaseUser:     os.Getenv("COUCHBASE_USER"),
		CouchbasePassword: os.Getenv("COUCHBASE_PASSWORD"),
		CouchbasePort:     os.Getenv("COUCHBASE_PORT"),
		CouchbaseHost:     os.Getenv("COUCHBASE_HOST"),
	}
	err := validateConfig(log, config)
	fatalIfErr(log, err)

	couchClusterResponses, getCouchClusterStatsError := getCouchClusterStats(log, config)
	couchBucketResponses, getCouchBucketStatsError := getCouchBucketsStats(log, config)
	couchReplicationResponses, getCouchReplicationStatsError := getCouchReplicationStats(log, config)
	couchRemoteReplicationResponses, getCouchRemoteReplicationStatsError := getCouchRemoteReplicationStats(log, config)
	for _, currentError := range []interface{}{getCouchClusterStatsError, getCouchBucketStatsError, getCouchReplicationStatsError, getCouchRemoteReplicationStatsError} {
		if getCouchClusterStatsError != nil {
			log.WithFields(logrus.Fields{
				"err": currentError,
			}).Fatal("Error retreiving couchbase stats.")
		}
	}

	data.Metrics = append(data.Metrics, couchClusterResponses...)
	data.Metrics = append(data.Metrics, couchBucketResponses...)
	data.Metrics = append(data.Metrics, couchReplicationResponses...)
	data.Metrics = append(data.Metrics, couchRemoteReplicationResponses...)
	fatalIfErr(log, helpers.OutputJSON(data, prettyPrint))
}

func avgInt64Sample(sampleSet []int64) (result float32) {
	var sampleSetLength = len(sampleSet)
	if sampleSetLength > 0 {
		var total int64
		for _, currentSample := range sampleSet {
			total += currentSample
		}
		return float32(total) / float32(sampleSetLength)
	}
	return 0
}

func avgFloat32Sample(sampleSet []float32) (result float32) {
	var sampleSetLength = len(sampleSet)
	if sampleSetLength > 0 {
		var total float32
		for _, currentSample := range sampleSet {
			total += currentSample
		}
		return total / float32(sampleSetLength)
	}
	return 0
}

func formatBucketInfoStatsStructToMap(completeBucketInfo CompleteBucketInfo) (bucketInfoMap map[string]interface{}) {
	return map[string]interface{}{
		"event_type":                                           EVENT_TYPE,
		"provider":                                             PROVIDER,
		"couchbase.scalr.clustername":                          os.Getenv("CB_CLUSTER_NAME"),
		"couchbase.by_bucket.name":                             completeBucketInfo.bucketInfo.Name,
		"couchbase.by_bucket.avg_bg_wait_time":                 avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.AVGBGWaitTime),
		"couchbase.by_bucket.avg_disk_commit_time":             avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.AVGDiskCommitTime),
		"couchbase.by_bucket.bytes_read":                       avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.BytesRead),
		"couchbase.by_bucket.bytes_written":                    avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.BytesWritten),
		"couchbase.by_bucket.cas_hits":                         avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.CasHits),
		"couchbase.by_bucket.cas_misses":                       avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.CasMisses),
		"couchbase.by_bucket.cmd_get":                          avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.CMDGet),
		"couchbase.by_bucket.cmd_set":                          avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.CMDSet),
		"couchbase.by_bucket.couch_docs_actual_disk_size":      avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.CouchDocsActualDiskSize),
		"couchbase.by_bucket.couch_docs_data_size":             avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.CouchDocsDataSize),
		"couchbase.by_bucket.couch_docs_disk_size":             avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.CouchDocsDiskSize),
		"couchbase.by_bucket.couch_docs_fragmentation":         avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.CouchDocsFragmentation),
		"couchbase.by_bucket.couch_total_disk_size":            avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.CouchTotalDiskSize),
		"couchbase.by_bucket.couch_views_fragmentation":        avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.CouchViewsFragmentation),
		"couchbase.by_bucket.couch_views_ops":                  avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.CouchViewsOps),
		"couchbase.by_bucket.cpu_idle_ms":                      avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.CPUIdleTime),
		"couchbase.by_bucket.cpu_utilization_rate":             avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.CPUUtilizationRate),
		"couchbase.by_bucket.curr_connections":                 avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.CurrConnections),
		"couchbase.by_bucket.curr_items":                       avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.CurrItems),
		"couchbase.by_bucket.curr_items_tot":                   avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.CurrItemsTotal),
		"couchbase.by_bucket.decr_hits":                        avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.DecrHits),
		"couchbase.by_bucket.decr_misses":                      avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.DecrMisses),
		"couchbase.by_bucket.delete_hits":                      avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.DeleteHits),
		"couchbase.by_bucket.delete_misses":                    avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.DeleteMisses),
		"couchbase.by_bucket.disk_commit_count":                avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.DiskCommitCount),
		"couchbase.by_bucket.disk_update_count":                avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.DiskUpdateCount),
		"couchbase.by_bucket.disk_write_queue":                 avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.DiskWriteQueue),
		"couchbase.by_bucket.evictions":                        avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.Evictions),
		"couchbase.by_bucket.get_hits":                         avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.GetHits),
		"couchbase.by_bucket.get_misses":                       avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.GetMisses),
		"couchbase.by_bucket.hit_ratio":                        avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.HitRatio),
		"couchbase.by_bucket.incr_hits":                        avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.IncrHits),
		"couchbase.by_bucket.mem_free":                         avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.MemFree),
		"couchbase.by_bucket.mem_actual_free":                  avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.MemActuallFree),
		"couchbase.by_bucket.mem_total":                        avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.MemTotal),
		"couchbase.by_bucket.mem_used":                         avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.MemUsed),
		"couchbase.by_bucket.mem_actual_used":                  avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.MemActuallUsed),
		"couchbase.by_bucket.misses":                           avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.Misses),
		"couchbase.by_bucket.ops":                              avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.Ops),
		"couchbase.by_bucket.vb_active_itm_memory":             avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.VBActiveItmMemory),
		"couchbase.by_bucket.vb_active_meta_data_memory":       avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.VBActiveMetaDataMemory),
		"couchbase.by_bucket.vb_active_num":                    avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.VBActiveNums),
		"couchbase.by_bucket.vb_active_queue_drain":            avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.VBActiveQueueDrain),
		"couchbase.by_bucket.vb_active_queue_size":             avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.VBActiveQueueSize),
		"couchbase.by_bucket.vb_active_resident_items_ratio":   avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.VBActiveResidentItemsRatio),
		"couchbase.by_bucket.vb_active_num_non_resident":       avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.VBActiveNumNonResident),
		"couchbase.by_bucket.vb_avg_total_queue_age":           avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.VBAvgTotalQueueAge),
		"couchbase.by_bucket.vb_pending_ops_create":            avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.VBPendingOpsCreate),
		"couchbase.by_bucket.vb_pending_queue_fill":            avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.VBPendingQueueFill),
		"couchbase.by_bucket.vb_replica_curr_items":            avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.VBReplicaCurrItems),
		"couchbase.by_bucket.vb_replica_itm_memory":            avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.VBReplicaItmMemory),
		"couchbase.by_bucket.vb_replica_meta_data_memory":      avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.VBReplicaMetaDataMemory),
		"couchbase.by_bucket.vb_replica_num":                   avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.VBReplicaNum),
		"couchbase.by_bucket.vb_replica_queue_size":            avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.VBReplicaQueueSize),
		"couchbase.by_bucket.xdc_ops":                          avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.XDCOPS),
		"couchbase.by_bucket.vb_replica_resident_items_ration": avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.VBReplicaResidentItemsRatio),
	}
}

func formatBucketInfoEPStatsStructToMap(completeBucketInfo CompleteBucketInfo) (bucketInfoMap map[string]interface{}) {
	return map[string]interface{}{
		"event_type":                                          EVENT_TYPE,
		"provider":                                            PROVIDER,
		"couchbase.scalr.clustername":                         os.Getenv("CB_CLUSTER_NAME"),
		"couchbase.by_bucket.name":                            completeBucketInfo.bucketInfo.Name,
		"couchbase.by_bucket.ep_bg_fetched":                   avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPBGFetched),
		"couchbase.by_bucket.ep_cache_miss_rate":              avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPCacheMissRate),
		"couchbase.by_bucket.ep_diskqueue_items":              avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPDiskQueueItems),
		"couchbase.by_bucket.ep_diskqueue_drain":              avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPDiskQueueDrain),
		"couchbase.by_bucket.ep_diskqueue_fill":               avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPDiskQueueFill),
		"couchbase.by_bucket.ep_flusher_todo":                 avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPFlusherTodo),
		"couchbase.by_bucket.ep_item_commit_failed":           avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EpItemCommitFailed),
		"couchbase.by_bucket.ep_max_size":                     avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.EPMaxSize),
		"couchbase.by_bucket.ep_mem_high_wat":                 avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.EPMemHighWater),
		"couchbase.by_bucket.ep_num_non_resident":             avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPNumNonResident),
		"couchbase.by_bucket.ep_meta_data_memory":             avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPMetaDataMemory),
		"couchbase.by_bucket.ep_num_value_ejects":             avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPNumValueEjects),
		"couchbase.by_bucket.ep_num_ops_get_meta":             avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPNumOpsGetMeta),
		"couchbase.by_bucket.ep_num_ops_set_meta":             avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPNumOpsSetMeta),
		"couchbase.by_bucket.ep_oom_errors":                   avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPOOMErrors),
		"couchbase.by_bucket.ep_ops_create":                   avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPOPSCreate),
		"couchbase.by_bucket.ep_ops_update":                   avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPOPSUpdate),
		"couchbase.by_bucket.ep_overhead":                     avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.EPOverhead),
		"couchbase.by_bucket.ep_queue_size":                   avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPQueueSize),
		"couchbase.by_bucket.ep_resident_items_rate":          avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPResidentItemsRate),
		"couchbase.by_bucket.ep_tap_replica_queue_drain":      avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPTapReplicaQueueDrain),
		"couchbase.by_bucket.ep_tap_total_queue_drain":        avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPTapTotalQueueDrain),
		"couchbase.by_bucket.ep_tap_total_queue_fill":         avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPTapTotalQueueFill),
		"couchbase.by_bucket.ep_tap_total_total_backlog_size": avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPTapTotalTotalBacklogSize),
		"couchbase.by_bucket.ep_tmp_oom_errors":               avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPTMPOOMErrors),
		"couchbase.by_bucket.ep_kv_size":                      avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.EPKVSize),
		"couchbase.by_bucket.ep_mem_low_wat":                  avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.EPMemLowWater),
		"couchbase.by_bucket.ep_dcp_replica_items_remaining":  avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.EPDcpReplicaItemsRemaining),
		"couchbase.by_bucket.ep_dcp_replica_items_sent":       avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPDcpReplicaItemsSent),
		"couchbase.by_bucket.ep_dcp_replica_total_bytes":      avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPDcpReplicaTotalBytes),
		"couchbase.by_bucket.ep_dcp_xdcr_items_remaining":     avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.EPDcpXDCRItemsRemaining),
		"couchbase.by_bucket.ep_dcp_xdcr_items_sent":          avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPDcpXDCRItemsSent),
		"couchbase.by_bucket.ep_dcp_xdcr_total_bytes":         avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPDcpXDCRTotalBytes),
		"couchbase.by_bucket.ep_dcp_views_items_remaining":    avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.EPDcpViewItemsRemaining),
		"couchbase.by_bucket.ep_dcp_2i_items_remaining":       avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.EPDcp2iItemsRemaining),
		"couchbase.by_bucket.ep_dcp_fts_items_remaining":      avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.EPDcpFtsItemsRemaining),
		"couchbase.by_bucket.ep_dcp_other_items_remaining":    avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.EPDcpOtherItemsRemaining),
	}
}

func getCouchBucketsStats(log *logrus.Logger, couchConfig CouchbaseConfig) (allBucketStats []MetricData, err error) {
	allBucketStatsInfos, err := getAllBucketsInfo(log, couchConfig)
	if err != nil {
		return []MetricData{}, err
	}
	var bucketCount = len(allBucketStatsInfos)
	bucketStatsResponses := make(chan CompleteBucketInfo, bucketCount)
	var wg sync.WaitGroup
	wg.Add(bucketCount)
	for _, currentBucket := range allBucketStatsInfos {
		go func(currentBucket CouchbaseBucketStatsURI) {
			defer wg.Done()
			bucketStats, err := getBucketStats(log, couchConfig, currentBucket.StatsObject.URI)
			if err != nil {
				log.WithFields(logrus.Fields{
					"currentBucket": currentBucket,
					"error":         err,
				}).Info("Error Retreiving bucket stats")
			} else {
				bucketStatsResponses <- CompleteBucketInfo{currentBucket, bucketStats}
			}
		}(currentBucket)
	}
	wg.Wait()
	close(bucketStatsResponses)
	for response := range bucketStatsResponses {
		bucketList = append(bucketList, response.bucketInfo.Name)
		allBucketStats = append(allBucketStats, formatBucketInfoStatsStructToMap(response))
		allBucketStats = append(allBucketStats, formatBucketInfoEPStatsStructToMap(response))
	}

	return allBucketStats, nil
}

func getBucketStats(log *logrus.Logger, config CouchbaseConfig, bucketURI string) (bucketStats CouchbaseBucketStats, err error) {
	couchbaseStatsURI := fmt.Sprintf("%v:%v%v%v", config.CouchbaseHost, config.CouchbasePort, bucketURI, "?zoom=minute")
	httpReq, err := http.NewRequest("GET", couchbaseStatsURI, bytes.NewBuffer([]byte("")))
	if err != nil {
		log.WithFields(logrus.Fields{
			"couchbaseStatsURI": couchbaseStatsURI,
			"error":             err,
		}).Error("Encountered error creating http.NewRequest")
		return CouchbaseBucketStats{}, err
	}
	httpReq.SetBasicAuth(config.CouchbaseUser, config.CouchbasePassword)
	err = executeAndDecode(log, *httpReq, &bucketStats)
	if err != nil {
		return CouchbaseBucketStats{}, err
	}
	return bucketStats, nil
}

func getAllBucketsInfo(log *logrus.Logger, config CouchbaseConfig) (bucketStatsInfos []CouchbaseBucketStatsURI, err error) {
	couchbaseStatsURI := fmt.Sprintf("%v:%v/%v", config.CouchbaseHost, config.CouchbasePort, "pools/default/buckets")
	httpReq, err := http.NewRequest("GET", couchbaseStatsURI, bytes.NewBuffer([]byte("")))
	if err != nil {
		log.WithFields(logrus.Fields{
			"couchbaseStatsURI": couchbaseStatsURI,
			"error":             err,
		}).Error("Encountered error creating http.NewRequest")
		return []CouchbaseBucketStatsURI{}, err
	}
	httpReq.SetBasicAuth(config.CouchbaseUser, config.CouchbasePassword)
	err = executeAndDecode(log, *httpReq, &bucketStatsInfos)
	if err != nil {
		return []CouchbaseBucketStatsURI{}, err
	}
	return bucketStatsInfos, nil
}

func getClusterInfo(log *logrus.Logger, config CouchbaseConfig) (clusterRecord CouchbaseClusterInfo, err error) {
	couchbaseStatsURI := fmt.Sprintf("%v:%v/%v", config.CouchbaseHost, config.CouchbasePort, "pools/default")
	httpReq, err := http.NewRequest("GET", couchbaseStatsURI, bytes.NewBuffer([]byte("")))
	if err != nil {
		log.WithFields(logrus.Fields{
			"couchbaseStatsURI": couchbaseStatsURI,
			"error":             err,
		}).Error("Encountered error creating http.NewRequest")
		return CouchbaseClusterInfo{}, err
	}
	httpReq.SetBasicAuth(config.CouchbaseUser, config.CouchbasePassword)
	err = executeAndDecode(log, *httpReq, &clusterRecord)
	if err != nil {
		return CouchbaseClusterInfo{}, err
	}
	return clusterRecord, nil
}

type CouchbaseIndexStatusResponse struct {
	Indexes []CouchbaseIndex
}

func getClusterIndexStatus(log *logrus.Logger, config CouchbaseConfig, indexStatusUrl string) ([]CouchbaseIndex, error) {
	couchbaseStatsURI := fmt.Sprintf("%v:%v/%v", config.CouchbaseHost, config.CouchbasePort, indexStatusUrl)
	httpReq, err := http.NewRequest("GET", couchbaseStatsURI, bytes.NewBuffer([]byte("")))
	if err != nil {
		log.WithFields(logrus.Fields{
			"couchbaseStatsURI": couchbaseStatsURI,
			"error":             err,
		}).Error("Encountered error creating http.NewRequest")
		return []CouchbaseIndex{}, err
	}
	httpReq.SetBasicAuth(config.CouchbaseUser, config.CouchbasePassword)
	var couchbaseIndexesResponse CouchbaseIndexStatusResponse
	err = executeAndDecode(log, *httpReq, &couchbaseIndexesResponse)
	if err != nil {
		return []CouchbaseIndex{}, err
	}
	return couchbaseIndexesResponse.Indexes, nil
}

func getCouchClusterStats(log *logrus.Logger, config CouchbaseConfig) ([]MetricData, error) {
	clusterResponse, err := getClusterInfo(log, config)
	if err != nil {
		log.WithFields(logrus.Fields{
			"CouchbaseConfig": config,
			"error":           err,
		}).Error("Encountered error querying Nodes")
		return make([]MetricData, 0), err
	}

	var returnMetrics []MetricData
	// add by node cluster metrics
	for _, node := range clusterResponse.Nodes {
		returnMetrics = append(returnMetrics,
			MetricData{
				"event_type":                                   EVENT_TYPE,
				"provider":                                     PROVIDER,
				"couchbase.cluster.name":                       clusterResponse.Name,
				"couchbase.cluster.by_node.status":             node.Status,
				"couchbase.cluster.by_node.hostname":           node.HostName,
				"couchbase.cluster.by_node.cluster_membership": node.ClusterMembership,
			},
		)
	}

	if clusterResponse.IndexStatusURI != "" {
		couchbaseIndexes, err := getClusterIndexStatus(log, config, clusterResponse.IndexStatusURI)
		if err != nil {
			log.WithFields(logrus.Fields{
				"CouchbaseConfig": config,
				"error":           err,
			}).Error("Encountered error querying Cluster Indexes")
		}
		for _, node := range couchbaseIndexes {
			returnMetrics = append(returnMetrics,
				MetricData{
					"event_type":                  "CouchbaseIndexSample",
					"provider":                    PROVIDER,
					"couchbase.scalr.clustername": os.Getenv("CB_CLUSTER_NAME"),
					"couchbase.index.id":          node.ID,
					"couchbase.index.index":       node.Index,
					"couchbase.index.definition":  node.Definition,
					"couchbase.index.status":      node.Status,
					"couchbase.index.progress":    node.Progress,
				},
			)
		}
	}

	// finally, add top level cluster metrics
	return append(returnMetrics,
		MetricData{
			"event_type":                         EVENT_TYPE,
			"provider":                           PROVIDER,
			"couchbase.scalr.clustername":        os.Getenv("CB_CLUSTER_NAME"),
			"couchbase.cluster.name":             clusterResponse.Name,
			"couchbase.cluster.hdd.free":         clusterResponse.StorageTotals.HDD.HDDFree,
			"couchbase.cluster.hdd.total":        clusterResponse.StorageTotals.HDD.HDDTotal,
			"couchbase.cluster.hdd.quota_total":  clusterResponse.StorageTotals.HDD.HDDQuotaTotal,
			"couchbase.cluster.hdd.used":         clusterResponse.StorageTotals.HDD.HDDUsed,
			"couchbase.cluster.hdd.used_by_data": clusterResponse.StorageTotals.HDD.HDDUsedByData,
			"couchbase.cluster.ram.total":        clusterResponse.StorageTotals.RAM.RAMTotal,
			"couchbase.cluster.ram.quota_total":  clusterResponse.StorageTotals.RAM.RAMQuotaTotal,
			"couchbase.cluster.ram.used":         clusterResponse.StorageTotals.RAM.RAMUsed,
			"couchbase.cluster.ram.used_by_data": clusterResponse.StorageTotals.RAM.RAMUsedByData,
		},
	), nil
}

type couchbaseReplicationStats struct {
	Hostname string `json:"hostname"`
	Name     string `json:"name"`
	URI      string `json:"uri"`
	Username string `json:"username"`
	UUID     string `json:"uuid"`
	Deleted  bool   `json:"deleted"`
}

func getCouchReplicationStats(log *logrus.Logger, config CouchbaseConfig) ([]MetricData, error) {
	couchbaseReplicationStatsURI := fmt.Sprintf("%v:%v/%v", config.CouchbaseHost, config.CouchbasePort, "pools/default/remoteClusters")
	httpReq, err := http.NewRequest("GET", couchbaseReplicationStatsURI, bytes.NewBuffer([]byte("")))
	returnMetrics := make([]MetricData, 0)
	if err != nil {
		log.WithFields(logrus.Fields{
			"couchbaseReplicationStatsURI": couchbaseReplicationStatsURI,
			"error": err,
		}).Error("Encountered error creating http.NewRequest")
		return returnMetrics, err
	}
	httpReq.SetBasicAuth(config.CouchbaseUser, config.CouchbasePassword)
	var replicationStats []couchbaseReplicationStats
	err = executeAndDecode(log, *httpReq, &replicationStats)
	if err != nil {
		return returnMetrics, err
	}

	// add by node cluster metrics
	for _, replication := range replicationStats {
		remoteUUIDList = append(remoteUUIDList, replication.UUID)
		returnMetrics = append(returnMetrics,
			MetricData{
				"event_type":                     "CouchbaseReplicationSample",
				"provider":                       PROVIDER,
				"couchbase.replication.hostname": replication.Hostname,
				"couchbase.replication.name":     replication.Name,
				"couchbase.replication.uri":      replication.URI,
				"couchbase.replication.username": replication.Username,
				"couchbase.replication.uuid":     replication.UUID,
				"couchbase.replication.deleted":  replication.Deleted,
			},
		)
	}

	return returnMetrics, nil
}

type remoteMeticChanResp struct {
	Data MetricData
	Err  error
}

func getCouchRemoteReplicationStats(log *logrus.Logger, config CouchbaseConfig) ([]MetricData, error) {
	returnMetrics := make([]MetricData, 0)
	statsChan := make(chan remoteMeticChanResp)
	wg := &sync.WaitGroup{}

	for _, bucket := range bucketList {
		for _, uuid := range remoteUUIDList {
			for _, endpoint := range remoteStatEndpoints {
				wg.Add(1)
				go processRemoteReplicationStats(log, config, wg, statsChan, bucket, uuid, endpoint)
			}
		}
	}

	go func() {
		wg.Wait()
		close(statsChan)
	}()

	for stat := range statsChan {
		if stat.Err == nil {
			returnMetrics = append(returnMetrics, stat.Data)
		}
	}

	return returnMetrics, nil
}

func processRemoteReplicationStats(log *logrus.Logger, config CouchbaseConfig, wg *sync.WaitGroup, statsChan chan<- remoteMeticChanResp, bucket string, uuid string, endpoint string) {
	defer wg.Done()
	encoded := fmt.Sprintf("%%2F%s%%2F%s%%2F%s%%2f%s", uuid, bucket, bucket, endpoint)
	uri := fmt.Sprintf("%s:%s/pools/default/buckets/%s/stats/replications%s", config.CouchbaseHost, config.CouchbasePort, bucket, encoded)
	httpReq, err := http.NewRequest("GET", uri, bytes.NewBuffer([]byte("")))
	if err != nil {
		log.WithFields(logrus.Fields{
			"uri":   uri,
			"error": err,
		}).Error("Encountered error creating http.NewRequest")
		statsChan <- remoteMeticChanResp{
			Data: MetricData{},
			Err:  err,
		}
	}
	httpReq.SetBasicAuth(config.CouchbaseUser, config.CouchbasePassword)

	stat := CouchbaseRemoteReplicationStats{}
	err = executeAndDecode(log, *httpReq, &stat)
	if err != nil {
		fmt.Println("would error")
		statsChan <- remoteMeticChanResp{
			Data: MetricData{},
			Err:  err,
		}
	}

	statsChan <- remoteMeticChanResp{
		Data: MetricData{
			"event_type": "CouchbaseReplicationSample",
			"provider":   PROVIDER,
			fmt.Sprintf("couchbase.replication.%s.samplescount", endpoint): stat.SamplesCount,
			fmt.Sprintf("couchbase.replication.%s.ispersistent", endpoint): stat.IsPersistent,
			fmt.Sprintf("couchbase.replication.%s.lasttstamp", endpoint):   stat.LastTStamp,
			fmt.Sprintf("couchbase.replication.%s.interval", endpoint):     stat.Internal,
			fmt.Sprintf("couchbase.replication.%s.timestamp", endpoint):    stat.Timestamp,
			fmt.Sprintf("couchbase.replication.%s.nodestats", endpoint):    stat.NodStats,
		},
		Err: err,
	}
}
