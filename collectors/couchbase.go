package collectors

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"
)

type CouchbaseBucketInfo struct {
	Name string `json:"name"`
}

//EPCacheMissRatio         []int   `json:"ep_cache_miss_ratio"`          //percent
type CouchbaseBucketStats struct {
	OP struct {
		Samples struct {
			AVGBGWaitTime              []int   `json:"avg_bg_wait_time"`                //seconds
			AVGDiskCommitTime          []int   `json:"avg_disk_commit_time"`            //seconds
			BytesRead                  []int64 `json:"bytes_read"`                      //bytes
			BytesWritten               []int64 `json:"bytes_written"`                   //bytes
			CasHits                    []int   `json:"cas_hits"`                        //hits
			CasMisses                  []int   `json:"cas_misses"`                      //misses
			CMDGet                     []int   `json:"cmd_get"`                         //gets
			CMDSet                     []int   `json:"cmd_set"`                         //sets
			CouchDocsActualDiskSize    []int   `json:"couch_docs_actual_disk_size"`     //bytes
			CouchDocsDataSize          []int64 `json:"couch_docs_data_size"`            //bytes
			CouchDocsDiskSize          []int64 `json:"couch_docs_disk_size"`            //bytes
			CouchDocsFragmentation     []int   `json:"couch_docs_fragmentation"`        //percent
			CouchTotalDiskSize         []int64 `json:"couch_total_disk_size"`           //bytes
			CouchViewsFragmentation    []int   `json:"couch_views_fragmentation"`       //percent
			CouchViewsOps              []int   `json:"couch_views_ops"`                 //operations
			CPUIdleTime                []int   `json:"cpu_idle_ms"`                     //milliseconds
			CPUUtilizationRate         []int   `json:"cpu_utilization_rate"`            //percent
			CurrConnections            []int   `json:"curr_connections"`                //connections
			CurrItems                  []int   `json:"curr_items"`                      //items
			CurrItemsTotal             []int   `json:"curr_items_tot"`                  //items
			DecrHits                   []int   `json:"decr_hits"`                       //hits
			DecrMisses                 []int   `json:"decr_misses"`                     //misses
			DeleteHits                 []int   `json:"delete_hits"`                     //hits
			DeleteMisses               []int   `json:"delete_misses"`                   //misses
			DiskCommitCount            []int   `json:"disk_commit_count"`               //operations
			DiskUpdateCount            []int   `json:"disk_update_count"`               //operations
			DiskWriteQueue             []int   `json:"disk_write_queue"`                //operations
			EPBGFetched                []int   `json:"ep_bg_fetched"`                   //fetchs/second
			EPCacheMissRate            []int   `json:"ep_cache_miss_rate"`              //misses
			EPDiskQueueDrain           []int   `json:"ep_diskqueue_drain"`              //items
			EPDiskQueueFill            []int   `json:"ep_diskqueue_fill"`               //items
			EPFlusherTodo              []int   `json:"ep_flusher_todo"`                 //items
			EpItemCommitFailed         []int   `json:"ep_item_commit_failed"`           //errors
			EPMaxSize                  []int64 `json:"ep_max_size"`                     //bytes
			EPMemHighWater             []int64 `json:"ep_mem_high_wat"`                 //bytes
			EPNumNonResident           []int64 `json:"ep_num_non_resident"`             //Items
			EPNumValueEjects           []int64 `json:"ep_num_value_ejects"`             //Items
			EPOOMErrors                []int   `json:"ep_oom_errors"`                   //errors
			EPOPSCreate                []int   `json:"ep_ops_create"`                   //operations
			EPOPSUpdate                []int   `json:"ep_ops_update"`                   //operations
			EPOverhead                 []int64 `json:"ep_overhead"`                     //bytes
			EPQueueSize                []int   `json:"ep_queue_size"`                   //items
			EPResidentItemsRate        []int   `json:"ep_resident_items_rate"`          //items
			EPTapReplicaQueueDrain     []int   `json:"ep_tap_replica_queue_drain"`      //items
			EPTapTotalQueueDrain       []int   `json:"ep_tap_total_queue_drain"`        //items
			EPTapTotalQueueFill        []int   `json:"ep_tap_total_queue_fill"`         //items
			EPTapTotalTotalBacklogSize []int   `json:"ep_tap_total_total_backlog_size"` //items
			EPTMPOOMErrors             []int   `json:"ep_tmp_oom_errors"`               //errors
			Evictions                  []int   `json:"evictions"`                       //evictions
			GetHits                    []int   `json:"get_hits"`                        //hits
			GetMisses                  []int   `json:"get_misses"`                      //misses
			HitRatio                   []int   `json:"hit_ratio"`                       //percent
			IncrHits                   []int   `json:"incr_hits"`                       //hits
			MemFree                    []int64 `json:"mem_free"`                        //bytes
			MemActuallFree             []int64 `json:"mem_actual_free"`                 //bytes
			MemTotal                   []int64 `json:"mem_total"`                       //bytes
			MemUsed                    []int64 `json:"mem_used"`                        //bytes
			MemActuallUsed             []int64 `json:"mem_actual_used"`                 //bytes
			Misses                     []int   `json:"misses"`                          //misses
			Ops                        []int   `json:"ops"`                             //operations
		} `json:"samples"`
	} `json:"op"`
}

type CouchbaseBucketStatsUri struct {
	URI         string `json:"uri"`
	StatsObject struct {
		URI string `json:"uri"`
	} `json:"stats"`
}

type CouchbaseClusterInfo struct {
	Name          string `json:"name"`
	StorageTotals struct {
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
}

func getBucketsStats(config CouchbaseConfig, bucketUri string) (bucketStats CouchbaseBucketStats, err error) {
	couchbaseStatsURI := fmt.Sprintf("%v:%v/%v", config.CouchbaseHost, config.CouchbasePort, bucketUri)
	httpReq, err := http.NewRequest("GET", couchbaseStatsURI, bytes.NewBuffer([]byte("")))
	if err != nil {
		log.WithFields(logrus.Fields{
			"couchbaseStatsURI": couchbaseStatsURI,
			"error":             err,
		}).Error("Encountered error creating http.NewRequest")
		return CouchbaseBucketStats{}, err
	}
	httpReq.SetBasicAuth(config.CouchbaseUser, config.CouchbasePassword)
	err = executeAndDecode(*httpReq, &bucketStats)
	if err != nil {
		return CouchbaseBucketStats{}, err
	}
	return bucketStats, nil
}

func getBucketsStatsUris(config CouchbaseConfig) (bucketStatsUris []CouchbaseBucketStatsUri, err error) {
	couchbaseStatsURI := fmt.Sprintf("%v:%v/%v", config.CouchbaseHost, config.CouchbasePort, "pools/default/buckets")
	httpReq, err := http.NewRequest("GET", couchbaseStatsURI, bytes.NewBuffer([]byte("")))
	if err != nil {
		log.WithFields(logrus.Fields{
			"couchbaseStatsURI": couchbaseStatsURI,
			"error":             err,
		}).Error("Encountered error creating http.NewRequest")
		return []CouchbaseBucketStatsUri{}, err
	}
	httpReq.SetBasicAuth(config.CouchbaseUser, config.CouchbasePassword)
	err = executeAndDecode(*httpReq, &bucketStatsUris)
	if err != nil {
		return []CouchbaseBucketStatsUri{}, err
	}
	return bucketStatsUris, nil
}

func getClusterInfo(config CouchbaseConfig) (clusterRecord CouchbaseClusterInfo, err error) {
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
	err = executeAndDecode(*httpReq, &clusterRecord)
	if err != nil {
		return CouchbaseClusterInfo{}, err
	}
	return clusterRecord, nil
}

func getCouchClusterStats(config CouchbaseConfig) ([]map[string]interface{}, error) {
	clusterResponse, err := getClusterInfo(config)
	if err != nil {
		log.WithFields(logrus.Fields{
			"CouchbaseConfig": config,
			"error":           err,
		}).Error("Encountered error querying Nodes")
		return make([]map[string]interface{}, 0), err
	}
	Stats := make([]map[string]interface{}, 0)
	Stats = append(Stats, map[string]interface{}{
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
	})
	return Stats, nil
}

//CouchbaseCollector gets the couch stats.
func CouchbaseCollector(config Config, stats chan<- []map[string]interface{}) {
	var couchConfig CouchbaseConfig
	err := mapstructure.Decode(config.Collectors["couchbase"].CollectorConfig, &couchConfig)
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("Unable to decode RabbitMq config into RabbitMQConfig object")
		close(stats)
	}
	couchResponses, getStatsError := getCouchClusterStats(couchConfig)
	if getStatsError != nil {
		log.WithFields(logrus.Fields{
			"err": getStatsError,
		}).Error("Error retreiving rabbitmq stats.")
		close(stats)
		return
	}
	stats <- couchResponses
}
