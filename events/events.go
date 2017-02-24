package events

import (
	"fmt"
	"math/rand"

	"github.com/GannettDigital/go-newrelic-plugin/metrics"
	"github.com/GannettDigital/paas-api-utils/utilsHTTP"
	"github.com/Sirupsen/logrus"
)

var log = logrus.New()

// NGINXEvent - record NGINX metrics
func (config *Config) NGINXEvent() {
	var runner utilsHTTP.HTTPRunnerImpl
	pollResult := metrics.PollStatus(&config.NGINXConfig, runner)

	log.WithFields(logrus.Fields{
		"config": config,
	}).Info("Reporting NGINXMetrics to NewRelic")

	config.App.RecordCustomEvent("NGINXMetrics", map[string]interface{}{
		"nginx.net.connections": pollResult.Connections,
		"nginx.net.accepts":     pollResult.Accepts,
		"nginx.net.handled":     pollResult.Handled,
		"nginx.net.requests":    pollResult.Requests,
		"nginx.net.writing":     pollResult.Writing,
		"nginx.net.waiting":     pollResult.Waiting,
		"nginx.net.reading":     pollResult.Reading,
	})
}

// CustomEvent comment here
func (config *Config) CustomEventExample() {
	customeventname := "goNginxText"
	rint := rand.Intn(100)
	cpu := metrics.GetCPULoad()
	mem := metrics.GetMemFree()

	fmt.Println("Testing Go Nginx Test")

	config.App.RecordCustomEvent(customeventname, map[string]interface{}{
		"testInt": rint,
		"testCPU": cpu,
		"testMem": mem,
	})
}
