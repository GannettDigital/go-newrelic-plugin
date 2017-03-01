package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	newrelicMonitoring "github.com/newrelic/go-agent"
)

func main() {
	// list of collectors that exist
	// the key needs to match the value as defined in the config file
	// the value is the collector method that will be used to gather the stats for that type
	collectorArray := map[string]Collector{
		"nginx": nginxCollector,
	}

	// TODO: populate config
	config := Config{}

	app := setupNewRelic(config)
	// main routine
	for name, collector := range collectorArray {
		go func(collectorName string, collectorValue Collector) {
			//if _, exists := config[collectorName]; exists {
			//if config[collectorName]["enabled"] == "true" || true {
			// TODO: random delay to offset collections
			// TODO: time sourced from config
			ticker := time.NewTicker(time.Millisecond * 500)
			for _ = range ticker.C {
				go getResult(collectorName, app, config, collectorValue)
			}
			//}
			//}
		}(name, collector) // you must close over this variable or it will change on the function when the next iteration occurs https://github.com/golang/go/wiki/CommonMistakes
	}

	done := make(chan bool)
	<-done // block forever

}

func getResult(collectorName string, app newrelicMonitoring.Application, config Config, collector Collector) {
	fmt.Println("get result for", collectorName)
	c := make(chan map[string]string, 1)
	collector(config, c)

	select {
	case res := <-c:
		sendData(collectorName, app, config, res)
	case <-time.After(time.Second * 10):
		// timeout so we don't leaving threads that block forever
	}
}

func sendData(collectorName string, app newrelicMonitoring.Application, config Config, stats map[string]string) {
	// send stats
	App.RecordCustomEvent(collectorName, stats)
}

func setupNewRelic(config Config) newrelicMonitoring.Application {

	// TODO: pull from config
	// Create an app config.  Application name and New Relic license key are required.
	cfg := newrelicMonitoring.NewConfig("test-newrelic-plugin", os.Getenv("NR_KEY"))

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
