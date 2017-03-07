package collectors

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"
)

//CouchbaseCollector gets the couch stats.
func CouchbaseCollector(config Config, stats chan<- []map[string]interface{}) {
	var couchConfig CouchbaseConfig
	err := decodeCouchbaseConfig(config, &couchConfig)
	if err != nil {
		close(stats)
	}
	couchClusterResponses, getCouchClusterStatsError := getCouchClusterStats(couchConfig)
	couchBucketResponses, getCouchBucketStatsError := getCouchBucketsStats(couchConfig)
	for _, currentError := range []interface{}{getCouchClusterStatsError, getCouchBucketStatsError} {
		if getCouchClusterStatsError != nil {
			log.WithFields(logrus.Fields{
				"err": currentError,
			}).Error("Error retreiving couchbase stats.")
			close(stats)
			return
		}
	}
	stats <- append(couchBucketResponses, couchClusterResponses...)
}

type CouchbaseBucketStats struct {
	OP struct {
		Samples struct {
			AVGBGWaitTime              []int     `json:"avg_bg_wait_time"`                //seconds
			AVGDiskCommitTime          []float32 `json:"avg_disk_commit_time"`            //seconds
			BytesRead                  []float32 `json:"bytes_read"`                      //bytes
			BytesWritten               []float32 `json:"bytes_written"`                   //bytes
			CasHits                    []int     `json:"cas_hits"`                        //hits
			CasMisses                  []int     `json:"cas_misses"`                      //misses
			CMDGet                     []int     `json:"cmd_get"`                         //gets
			CMDSet                     []int     `json:"cmd_set"`                         //sets
			CouchDocsActualDiskSize    []int     `json:"couch_docs_actual_disk_size"`     //bytes
			CouchDocsDataSize          []int64   `json:"couch_docs_data_size"`            //bytes
			CouchDocsDiskSize          []int64   `json:"couch_docs_disk_size"`            //bytes
			CouchDocsFragmentation     []int     `json:"couch_docs_fragmentation"`        //percent
			CouchTotalDiskSize         []int64   `json:"couch_total_disk_size"`           //bytes
			CouchViewsFragmentation    []int     `json:"couch_views_fragmentation"`       //percent
			CouchViewsOps              []int     `json:"couch_views_ops"`                 //operations
			CPUIdleTime                []int     `json:"cpu_idle_ms"`                     //milliseconds
			CPUUtilizationRate         []float32 `json:"cpu_utilization_rate"`            //percent
			CurrConnections            []int     `json:"curr_connections"`                //connections
			CurrItems                  []int     `json:"curr_items"`                      //items
			CurrItemsTotal             []int     `json:"curr_items_tot"`                  //items
			DecrHits                   []int     `json:"decr_hits"`                       //hits
			DecrMisses                 []int     `json:"decr_misses"`                     //misses
			DeleteHits                 []int     `json:"delete_hits"`                     //hits
			DeleteMisses               []int     `json:"delete_misses"`                   //misses
			DiskCommitCount            []float32 `json:"disk_commit_count"`               //operations
			DiskUpdateCount            []int     `json:"disk_update_count"`               //operations
			DiskWriteQueue             []int     `json:"disk_write_queue"`                //operations
			Evictions                  []int     `json:"evictions"`                       //evictions
			GetHits                    []int     `json:"get_hits"`                        //hits
			GetMisses                  []int     `json:"get_misses"`                      //misses
			HitRatio                   []int     `json:"hit_ratio"`                       //percent
			IncrHits                   []int     `json:"incr_hits"`                       //hits
			MemFree                    []int64   `json:"mem_free"`                        //bytes
			MemActuallFree             []int64   `json:"mem_actual_free"`                 //bytes
			MemTotal                   []int64   `json:"mem_total"`                       //bytes
			MemUsed                    []int64   `json:"mem_used"`                        //bytes
			MemActuallUsed             []int64   `json:"mem_actual_used"`                 //bytes
			Misses                     []int     `json:"misses"`                          //misses
			Ops                        []int     `json:"ops"`                             //operations
			VBActiveNums               []int     `json:"vb_active_num"`                   //items
			VBActiveQueueDrain         []int     `json:"vb_active_queue_drain"`           //items
			VBActiveQueueSize          []int     `json:"vb_active_queue_size"`            //items
			VBActiveResidentItemsRatio []float32 `json:"vb_active_resident_items_ratio"`  //items
			VBActiveNumNonResident     []int     `json:"vb_active_num_non_resident"`      //items
			VBAvgTotalQueueAge         []int     `json:"vb_avg_total_queue_age"`          //Seconds
			VBPendingOpsCreate         []int     `json:"vb_pending_ops_create"`           //operations
			VBPendingQueueFill         []int     `json:"vb_pending_queue_fill"`           //items
			VBReplicaCurrItems         []int     `json:"vb_replica_curr_items"`           //items
			VBReplicaMetaDataMemory    []int64   `json:"vb_replica_meta_data_memory"`     //bytes
			VBReplicaNum               []int     `json:"vb_replica_num"`                  //items
			VBReplicaQueueSize         []int     `json:"vb_replica_queue_size"`           //items
			XDCOPS                     []int     `json:"xdc_ops"`                         //operations
			EPBGFetched                []int     `json:"ep_bg_fetched"`                   //fetchs/second
			EPCacheMissRate            []int     `json:"ep_cache_miss_rate"`              //misses
			EPDiskQueueItems           []int     `json:"ep_diskqueue_items"`              //items
			EPDiskQueueDrain           []int     `json:"ep_diskqueue_drain"`              //items
			EPDiskQueueFill            []int     `json:"ep_diskqueue_fill"`               //items
			EPFlusherTodo              []int     `json:"ep_flusher_todo"`                 //items
			EpItemCommitFailed         []int     `json:"ep_item_commit_failed"`           //errors
			EPMaxSize                  []int64   `json:"ep_max_size"`                     //bytes
			EPMemHighWater             []int64   `json:"ep_mem_high_wat"`                 //bytes
			EPNumNonResident           []int64   `json:"ep_num_non_resident"`             //Items
			EPNumValueEjects           []int64   `json:"ep_num_value_ejects"`             //Items
			EPOOMErrors                []int     `json:"ep_oom_errors"`                   //errors
			EPOPSCreate                []int     `json:"ep_ops_create"`                   //operations
			EPOPSUpdate                []int     `json:"ep_ops_update"`                   //operations
			EPOverhead                 []int64   `json:"ep_overhead"`                     //bytes
			EPQueueSize                []int     `json:"ep_queue_size"`                   //items
			EPResidentItemsRate        []float32 `json:"ep_resident_items_rate"`          //items
			EPTapReplicaQueueDrain     []int     `json:"ep_tap_replica_queue_drain"`      //items
			EPTapTotalQueueDrain       []int     `json:"ep_tap_total_queue_drain"`        //items
			EPTapTotalQueueFill        []int     `json:"ep_tap_total_queue_fill"`         //items
			EPTapTotalTotalBacklogSize []int     `json:"ep_tap_total_total_backlog_size"` //items
			EPTMPOOMErrors             []int     `json:"ep_tmp_oom_errors"`               //errors
		} `json:"samples"`
	} `json:"op"`
}

type CouchbaseBucketStatsUri struct {
	Name        string `json:"name"`
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

type CompleteBucketInfo struct {
	bucketInfo  CouchbaseBucketStatsUri
	bucketStats CouchbaseBucketStats
}

func avgIntSample(sampleSet []int) (result float32) {
	var sampleSetLength int = len(sampleSet)
	if sampleSetLength > 0 {
		var total int = 0
		for _, currentSample := range sampleSet {
			total += currentSample
		}
		return float32(total) / float32(sampleSetLength)
	} else {
		return 0
	}
}

func avgInt64Sample(sampleSet []int64) (result float32) {
	var sampleSetLength int = len(sampleSet)
	if sampleSetLength > 0 {
		var total int64 = 0
		for _, currentSample := range sampleSet {
			total += currentSample
		}
		return float32(total) / float32(sampleSetLength)
	} else {
		return 0
	}
}

func avgFloat32Sample(sampleSet []float32) (result float32) {
	var sampleSetLength int = len(sampleSet)
	if sampleSetLength > 0 {
		var total float32 = 0
		for _, currentSample := range sampleSet {
			total += currentSample
		}
		return total / float32(sampleSetLength)
	} else {
		return 0
	}
}

func formatBucketInfoStatsStructToMap(completeBucketInfo CompleteBucketInfo) (bucketInfoMap map[string]interface{}) {
	return map[string]interface{}{
		"couchbase.by_bucket.name":                           completeBucketInfo.bucketInfo.Name,
		"couchbase.by_bucket.avg_bg_wait_time":               avgIntSample(completeBucketInfo.bucketStats.OP.Samples.AVGBGWaitTime),
		"couchbase.by_bucket.avg_disk_commit_time":           avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.AVGDiskCommitTime),
		"couchbase.by_bucket.bytes_read":                     avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.BytesRead),
		"couchbase.by_bucket.bytes_written":                  avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.BytesWritten),
		"couchbase.by_bucket.cas_hits":                       avgIntSample(completeBucketInfo.bucketStats.OP.Samples.CasHits),
		"couchbase.by_bucket.cas_misses":                     avgIntSample(completeBucketInfo.bucketStats.OP.Samples.CasMisses),
		"couchbase.by_bucket.cmd_get":                        avgIntSample(completeBucketInfo.bucketStats.OP.Samples.CMDGet),
		"couchbase.by_bucket.cmd_set":                        avgIntSample(completeBucketInfo.bucketStats.OP.Samples.CMDSet),
		"couchbase.by_bucket.couch_docs_actual_disk_size":    avgIntSample(completeBucketInfo.bucketStats.OP.Samples.CouchDocsActualDiskSize),
		"couchbase.by_bucket.couch_docs_data_size":           avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.CouchDocsDataSize),
		"couchbase.by_bucket.couch_docs_disk_size":           avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.CouchDocsDiskSize),
		"couchbase.by_bucket.couch_docs_fragmentation":       avgIntSample(completeBucketInfo.bucketStats.OP.Samples.CouchDocsFragmentation),
		"couchbase.by_bucket.couch_total_disk_size":          avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.CouchTotalDiskSize),
		"couchbase.by_bucket.couch_views_fragmentation":      avgIntSample(completeBucketInfo.bucketStats.OP.Samples.CouchViewsFragmentation),
		"couchbase.by_bucket.couch_views_ops":                avgIntSample(completeBucketInfo.bucketStats.OP.Samples.CouchViewsOps),
		"couchbase.by_bucket.cpu_idle_ms":                    avgIntSample(completeBucketInfo.bucketStats.OP.Samples.CPUIdleTime),
		"couchbase.by_bucket.cpu_utilization_rate":           avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.CPUUtilizationRate),
		"couchbase.by_bucket.curr_connections":               avgIntSample(completeBucketInfo.bucketStats.OP.Samples.CurrConnections),
		"couchbase.by_bucket.curr_items":                     avgIntSample(completeBucketInfo.bucketStats.OP.Samples.CurrItems),
		"couchbase.by_bucket.curr_items_tot":                 avgIntSample(completeBucketInfo.bucketStats.OP.Samples.CurrItemsTotal),
		"couchbase.by_bucket.decr_hits":                      avgIntSample(completeBucketInfo.bucketStats.OP.Samples.DecrHits),
		"couchbase.by_bucket.decr_misses":                    avgIntSample(completeBucketInfo.bucketStats.OP.Samples.DecrMisses),
		"couchbase.by_bucket.delete_hits":                    avgIntSample(completeBucketInfo.bucketStats.OP.Samples.DeleteHits),
		"couchbase.by_bucket.delete_misses":                  avgIntSample(completeBucketInfo.bucketStats.OP.Samples.DeleteMisses),
		"couchbase.by_bucket.disk_commit_count":              avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.DiskCommitCount),
		"couchbase.by_bucket.disk_update_count":              avgIntSample(completeBucketInfo.bucketStats.OP.Samples.DiskUpdateCount),
		"couchbase.by_bucket.disk_write_queue":               avgIntSample(completeBucketInfo.bucketStats.OP.Samples.DiskWriteQueue),
		"couchbase.by_bucket.evictions":                      avgIntSample(completeBucketInfo.bucketStats.OP.Samples.Evictions),
		"couchbase.by_bucket.get_hits":                       avgIntSample(completeBucketInfo.bucketStats.OP.Samples.GetHits),
		"couchbase.by_bucket.get_misses":                     avgIntSample(completeBucketInfo.bucketStats.OP.Samples.GetMisses),
		"couchbase.by_bucket.hit_ratio":                      avgIntSample(completeBucketInfo.bucketStats.OP.Samples.HitRatio),
		"couchbase.by_bucket.incr_hits":                      avgIntSample(completeBucketInfo.bucketStats.OP.Samples.IncrHits),
		"couchbase.by_bucket.mem_free":                       avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.MemFree),
		"couchbase.by_bucket.mem_actual_free":                avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.MemActuallFree),
		"couchbase.by_bucket.mem_total":                      avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.MemTotal),
		"couchbase.by_bucket.mem_used":                       avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.MemUsed),
		"couchbase.by_bucket.mem_actual_used":                avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.MemActuallUsed),
		"couchbase.by_bucket.misses":                         avgIntSample(completeBucketInfo.bucketStats.OP.Samples.Misses),
		"couchbase.by_bucket.ops":                            avgIntSample(completeBucketInfo.bucketStats.OP.Samples.Ops),
		"couchbase.by_bucket.vb_active_num":                  avgIntSample(completeBucketInfo.bucketStats.OP.Samples.VBActiveNums),
		"couchbase.by_bucket.vb_active_queue_drain":          avgIntSample(completeBucketInfo.bucketStats.OP.Samples.VBActiveQueueDrain),
		"couchbase.by_bucket.vb_active_queue_size":           avgIntSample(completeBucketInfo.bucketStats.OP.Samples.VBActiveQueueSize),
		"couchbase.by_bucket.vb_active_resident_items_ratio": avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.VBActiveResidentItemsRatio),
		"couchbase.by_bucket.vb_active_num_non_resident":     avgIntSample(completeBucketInfo.bucketStats.OP.Samples.VBActiveNumNonResident),
		"couchbase.by_bucket.vb_avg_total_queue_age":         avgIntSample(completeBucketInfo.bucketStats.OP.Samples.VBAvgTotalQueueAge),
		"couchbase.by_bucket.vb_pending_ops_create":          avgIntSample(completeBucketInfo.bucketStats.OP.Samples.VBPendingOpsCreate),
		"couchbase.by_bucket.vb_pending_queue_fill":          avgIntSample(completeBucketInfo.bucketStats.OP.Samples.VBPendingQueueFill),
		"couchbase.by_bucket.vb_replica_curr_items":          avgIntSample(completeBucketInfo.bucketStats.OP.Samples.VBReplicaCurrItems),
		"couchbase.by_bucket.vb_replica_meta_data_memory":    avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.VBReplicaMetaDataMemory),
		"couchbase.by_bucket.vb_replica_num":                 avgIntSample(completeBucketInfo.bucketStats.OP.Samples.VBReplicaNum),
		"couchbase.by_bucket.vb_replica_queue_size":          avgIntSample(completeBucketInfo.bucketStats.OP.Samples.VBReplicaQueueSize),
		"couchbase.by_bucket.xdc_ops":                        avgIntSample(completeBucketInfo.bucketStats.OP.Samples.XDCOPS),
	}
}

func formatBucketInfoEPStatsStructToMap(completeBucketInfo CompleteBucketInfo) (bucketInfoMap map[string]interface{}) {
	return map[string]interface{}{
		"couchbase.by_bucket.name":                            completeBucketInfo.bucketInfo.Name,
		"couchbase.by_bucket.ep_bg_fetched":                   avgIntSample(completeBucketInfo.bucketStats.OP.Samples.EPBGFetched),
		"couchbase.by_bucket.ep_cache_miss_rate":              avgIntSample(completeBucketInfo.bucketStats.OP.Samples.EPCacheMissRate),
		"couchbase.by_bucket.ep_diskqueue_items":              avgIntSample(completeBucketInfo.bucketStats.OP.Samples.EPDiskQueueItems),
		"couchbase.by_bucket.ep_diskqueue_drain":              avgIntSample(completeBucketInfo.bucketStats.OP.Samples.EPDiskQueueDrain),
		"couchbase.by_bucket.ep_diskqueue_fill":               avgIntSample(completeBucketInfo.bucketStats.OP.Samples.EPDiskQueueFill),
		"couchbase.by_bucket.ep_flusher_todo":                 avgIntSample(completeBucketInfo.bucketStats.OP.Samples.EPFlusherTodo),
		"couchbase.by_bucket.ep_item_commit_failed":           avgIntSample(completeBucketInfo.bucketStats.OP.Samples.EpItemCommitFailed),
		"couchbase.by_bucket.ep_max_size":                     avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.EPMaxSize),
		"couchbase.by_bucket.ep_mem_high_wat":                 avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.EPMemHighWater),
		"couchbase.by_bucket.ep_num_non_resident":             avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.EPNumNonResident),
		"couchbase.by_bucket.ep_num_value_ejects":             avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.EPNumValueEjects),
		"couchbase.by_bucket.ep_oom_errors":                   avgIntSample(completeBucketInfo.bucketStats.OP.Samples.EPOOMErrors),
		"couchbase.by_bucket.ep_ops_create":                   avgIntSample(completeBucketInfo.bucketStats.OP.Samples.EPOPSCreate),
		"couchbase.by_bucket.ep_ops_update":                   avgIntSample(completeBucketInfo.bucketStats.OP.Samples.EPOPSUpdate),
		"couchbase.by_bucket.ep_overhead":                     avgInt64Sample(completeBucketInfo.bucketStats.OP.Samples.EPOverhead),
		"couchbase.by_bucket.ep_queue_size":                   avgIntSample(completeBucketInfo.bucketStats.OP.Samples.EPQueueSize),
		"couchbase.by_bucket.ep_resident_items_rate":          avgFloat32Sample(completeBucketInfo.bucketStats.OP.Samples.EPResidentItemsRate),
		"couchbase.by_bucket.ep_tap_replica_queue_drain":      avgIntSample(completeBucketInfo.bucketStats.OP.Samples.EPTapReplicaQueueDrain),
		"couchbase.by_bucket.ep_tap_total_queue_drain":        avgIntSample(completeBucketInfo.bucketStats.OP.Samples.EPTapTotalQueueDrain),
		"couchbase.by_bucket.ep_tap_total_queue_fill":         avgIntSample(completeBucketInfo.bucketStats.OP.Samples.EPTapTotalQueueFill),
		"couchbase.by_bucket.ep_tap_total_total_backlog_size": avgIntSample(completeBucketInfo.bucketStats.OP.Samples.EPTapTotalTotalBacklogSize),
		"couchbase.by_bucket.ep_tmp_oom_errors":               avgIntSample(completeBucketInfo.bucketStats.OP.Samples.EPTMPOOMErrors),
	}
}

func getCouchBucketsStats(couchConfig CouchbaseConfig) (allBucketStats []map[string]interface{}, err error) {
	allBucketStatsInfos, err := getAllBucketsInfo(couchConfig)
	if err != nil {
		return []map[string]interface{}{}, err
	}
	var bucketCount = len(allBucketStatsInfos)
	bucketStatsResponses := make(chan CompleteBucketInfo, bucketCount)
	var wg sync.WaitGroup
	wg.Add(bucketCount)
	for _, currentBucket := range allBucketStatsInfos {
		go func(currentBucket CouchbaseBucketStatsUri) {
			defer wg.Done()
			bucketStats, err := getBucketStats(couchConfig, currentBucket.StatsObject.URI)
			if err != nil {
				log.WithFields(logrus.Fields{
					"currentBucket": currentBucket,
					"error":         err,
				}).Error("Error Retreiving bucket stats")
			} else {
				bucketStatsResponses <- CompleteBucketInfo{currentBucket, bucketStats}
			}
		}(currentBucket)
	}
	wg.Wait()
	close(bucketStatsResponses)
	for response := range bucketStatsResponses {
		allBucketStats = append(allBucketStats, formatBucketInfoStatsStructToMap(response))
		allBucketStats = append(allBucketStats, formatBucketInfoEPStatsStructToMap(response))
	}

	return allBucketStats, nil
}

func getBucketStats(config CouchbaseConfig, bucketUri string) (bucketStats CouchbaseBucketStats, err error) {
	couchbaseStatsURI := fmt.Sprintf("%v:%v%v%v", config.CouchbaseHost, config.CouchbasePort, bucketUri, "?zoom=minute")
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

func getAllBucketsInfo(config CouchbaseConfig) (bucketStatsInfos []CouchbaseBucketStatsUri, err error) {
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
	err = executeAndDecode(*httpReq, &bucketStatsInfos)
	if err != nil {
		return []CouchbaseBucketStatsUri{}, err
	}
	return bucketStatsInfos, nil
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

func decodeCouchbaseConfig(config Config, couchConfig *CouchbaseConfig) (err error) {
	err = mapstructure.Decode(config.Collectors["couchbase"].CollectorConfig, &couchConfig)
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("Unable to decode couchbase config into")
	}
	return err
}
