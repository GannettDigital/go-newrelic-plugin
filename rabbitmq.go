package goNewRelicCollector

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/GannettDigital/paas-api-utils/utilsHTTP"
	"github.com/Sirupsen/logrus"
)

type RabbitmqNodeMetrics struct {
	fd_used      int //Used File Descriptors
	mem_used     int //memory used in bytes
	partitions   int //Number of network partitions this node is seeing
	run_queue    int //Average Number of Eralng Processes waiting to run
	sockets_used int // Number of file descriptors used as sockets
}

type RabbitmqQueueMetrics struct {
	todo int //Queue specific metrics
}

var log = logrus.New()

func getRabbitmqStatus(config, runner utilsHTTP.HTTPRunner) string {
	statusUri := fmt.Sprintf("%v:%v/%v", config.host, config.port, config.endpoint)
	httpReq, err := http.NewRequest("GET", statusUri, bytes.NewBuffer([]byte("")))
	// http.NewRequest error
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Encountered error creating http.NewRequest")

		return ""
	}

	code, data, err := runner.CallAPI(log, nil, httpReq, &http.Client{})
	if err != nil || code != 200 {
		log.Error("Encountered error calling CallAPI")

		return ""
	}
	//Parse data...Not sure what it looks like yet.
	return string(data)
}

func scrapeStatus(status string) RabbitmqNodeMetrics {
	return RabbitmqNodeMetrics{
		fd_used:      1,
		mem_used:     2048,
		partitions:   3,
		run_queue:    40,
		sockets_used: 500,
	}
}

// TODO:
func rabbitmqCollector(config Config, stats chan<- map[string]string) {

	rabbitmqStatsMap := scrapeStatus(getRabbitmqStatus(config, runner))
	stats <- rabbitmqStatsMap
	/*
	  When called, this needs to:
	    1. collect metrics from rabbit
	    2. format metris into a map[string]interface{}
	    3. send that map[string]interface{} back on the channel (where the dispatcher will push it to NR)
	    4. thats all!
	*/
}
