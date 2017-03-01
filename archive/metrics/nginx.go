package metrics

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/GannettDigital/go-newrelic-plugin/helpers"
	"github.com/GannettDigital/paas-api-utils/utilsHTTP"
	"github.com/Sirupsen/logrus"
)

/*
Scrapes nginx_status page of the following format:

Active connections: 291
server accepts handled requests
 16630948 16630948 31070465
Reading: 6 Writing: 179 Waiting: 106
*/

var log = logrus.New()

func PollStatus(config *helpers.NginxConfig, runner utilsHTTP.HTTPRunner) helpers.NginxMetrics {
	return scrapeStatus(getNginxStatus(config, runner))
}

func getNginxStatus(config *helpers.NginxConfig, runner utilsHTTP.HTTPRunner) string {
	nginxStatus := fmt.Sprintf("%v:%v/%v", config.NginxStatusPage, config.NginxListenPort, config.NginxStatusURI)
	httpReq, err := http.NewRequest("GET", nginxStatus, bytes.NewBuffer([]byte("")))
	// http.NewRequest error
	if err != nil {
		log.WithFields(logrus.Fields{
			"nginxStatus": nginxStatus,
			"error":       err,
		}).Error("Encountered error creating http.NewRequest")

		return ""
	}

	code, data, err := runner.CallAPI(log, nil, httpReq, &http.Client{})
	if err != nil || code != 200 {
		log.WithFields(logrus.Fields{
			"code":                   code,
			"data":                   string(data),
			"httpReq":                httpReq,
			"config.NginxStatusPage": config.NginxStatusPage,
			"config.NginxListenPort": config.NginxListenPort,
			"config.NginxStatusURI":  config.NginxStatusURI,
			"error":                  err,
		}).Error("Encountered error calling CallAPI")

		return ""
	}

	return string(data)
}

func scrapeStatus(status string) helpers.NginxMetrics {
	multi := regexp.MustCompile(`Active connections: (\d+)`).FindString(status)
	contents := strings.Fields(multi)
	active := contents[2]

	multi = regexp.MustCompile(`(\d+)\s(\d+)\s(\d+)`).FindString(status)
	contents = strings.Fields(multi)
	accepts := contents[0]
	handled := contents[1]
	requests := contents[2]

	multi = regexp.MustCompile(`Reading: (\d+)`).FindString(status)
	contents = strings.Fields(multi)
	reading := contents[1]

	multi = regexp.MustCompile(`Writing: (\d+)`).FindString(status)
	contents = strings.Fields(multi)
	writing := contents[1]

	multi = regexp.MustCompile(`Waiting: (\d+)`).FindString(status)
	contents = strings.Fields(multi)
	waiting := contents[1]

	log.WithFields(logrus.Fields{
		"active":   active,
		"accepts":  accepts,
		"handled":  handled,
		"requests": requests,
		"reading":  reading,
		"writing":  writing,
		"waiting":  waiting,
	}).Warn("Encountered error creating http.NewRequest")

	return helpers.NginxMetrics{
		Connections: toInt(active),
		Accepts:     toInt(accepts),
		Handled:     toInt(handled),
		Requests:    toInt(requests),
		Writing:     toInt(writing),
		Waiting:     toInt(waiting),
		Reading:     toInt(reading),
	}
}

func toInt(value string) int {
	if value == "" {
		return 0
	} else {
		valueInt, err := strconv.Atoi(value)
		if err != nil {
			log.WithFields(logrus.Fields{
				"valueInt": valueInt,
				"error":    err,
			}).Error("Error converting value to int")

			return 0
		}

		return valueInt
	}
}
