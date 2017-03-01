package main

import (
	"fmt"
	"os"
	"time"

	"github.com/GannettDigital/go-newrelic-plugin/collectors"
	"github.com/Sirupsen/logrus"
	newrelicMonitoring "github.com/newrelic/go-agent"
)

var log = logrus.New()

func main() {

	// TODO: populate config from config.yaml
	config := collectors.Config{
		AppName: "test-newrelic-plugin",
		NginxConfig: collectors.NginxConfig{
			Enabled:         true,
			NginxListenPort: "8140",
			NginxStatusURI:  "nginx_status",
			NginxStatusPage: "http://localhost",
		},
	}

	app := setupNewRelic(config)

	// main routine
	for name, collector := range collectors.CollectorArray {
		go func(collectorName string, collectorValue collectors.Collector) {
			// TODO: random delay to offset collections
			// TODO: time sourced from config
			ticker := time.NewTicker(time.Millisecond * 1000)
			for _ = range ticker.C {
				go getResult(collectorName, app, config, collectorValue)
			}
		}(name, collector) // you must close over this variable or it will change on the function when the next iteration occurs https://github.com/golang/go/wiki/CommonMistakes
	}

	done := make(chan bool)
	<-done // block forever

}

func getResult(collectorName string, app newrelicMonitoring.Application, config collectors.Config, collector collectors.Collector) {
	c := make(chan map[string]interface{}, 1)
	collector(config, c)

	select {
	case res, success := <-c:
		if success {
			log.WithFields(logrus.Fields{
				"collector": collectorName,
			}).Info("received data from collector")
			sendData(collectorName, app, config, res)
		} else {
			log.WithFields(logrus.Fields{
				"collector": collectorName,
			}).Error("received error from collector")
		}
	case <-time.After(time.Second * 10):
		// timeout so we don't leaving threads that block forever
		log.WithFields(logrus.Fields{
			"collector": collectorName,
		}).Error("dispatcher timed out waiting for response from collector")
	}
}

func sendData(collectorName string, app newrelicMonitoring.Application, config collectors.Config, stats map[string]interface{}) {
	log.WithFields(logrus.Fields{
		"collector": collectorName,
	}).Info("recording event")
	// send stats
	app.RecordCustomEvent(fmt.Sprintf("gannettNewRelic%s", collectorName), stats)
}

func setupNewRelic(config collectors.Config) newrelicMonitoring.Application {

	// TODO: pull from config
	// Create an app config.  Application name and New Relic license key are required.
	cfg := newrelicMonitoring.NewConfig(config.AppName, os.Getenv("NR_KEY"))

	// Enable Go runtime metrics for the plugin
	cfg.RuntimeSampler.Enabled = true
	// Turn off unecessary transaction events since only custom events will be sent
	cfg.TransactionEvents.Enabled = false
	cfg.TransactionTracer.Enabled = false
	// Log to standard out.  Systemd will handle logging to journald
	cfg.Logger = newrelicMonitoring.NewDebugLogger(os.Stdout)

	// Create an application.  This represents an application in the New
	// Relic UI.
	app, err := newrelicMonitoring.NewApplication(cfg)
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("NewRelic application setup error")

		os.Exit(1)
	}

	if err := app.WaitForConnection(10 * time.Second); nil != err {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("Connection error")
	}

	return app
}
