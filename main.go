package main

import (
	"fmt"
	"os"
	"time"

	reflections "gopkg.in/oleiade/reflections.v1"

	"github.com/GannettDigital/go-newrelic-plugin/collectors"
	"github.com/Sirupsen/logrus"
	newrelicMonitoring "github.com/newrelic/go-agent"
)

var log = logrus.New()

func main() {

	// TODO: populate config from config.yaml
	config := collectors.Config{
		AppName: "test-newrelic-plugin",
		Nginx: collectors.NginxConfig{
			Enabled:         true,
			NginxListenPort: "8140",
			NginxStatusURI:  "nginx_status",
			NginxStatusPage: "http://localhost",
			PollIntervalMS:  500,
		},
	}

	app := setupNewRelic(config)

	// main routine
	for name, collector := range collectors.CollectorArray {
		conf, err := reflections.GetField(config, name)

		if err == nil {
			enabled, err := reflections.GetField(conf, "Enabled")
			if err == nil && enabled.(bool) {
				log.WithFields(logrus.Fields{
					"collector": name,
				}).Info("collector enabled, starting...")
				go func(collectorName string, collectorValue collectors.Collector) {

					// TODO: random delay to offset collections
					poll, _ := reflections.GetField(conf, "PollIntervalMS")
					ticker := time.NewTicker(time.Millisecond * time.Duration(poll.(int)))
					for _ = range ticker.C {
						go func() {
							defer func() {
								// recover from panic if one occured. Set err to nil otherwise.
								if err := recover(); err != nil {
									log.WithFields(logrus.Fields{
										"error": err,
									}).Error("collector panic'd, bad collector..")
								}
							}()
							getResult(collectorName, app, config, collectorValue)
						}()
					}
				}(name, collector) // you must close over this variable or it will change on the function when the next iteration occurs https://github.com/golang/go/wiki/CommonMistakes
			} else {
				log.WithFields(logrus.Fields{
					"collector": name,
				}).Info("collector not enabled")
			}
		} else {
			log.WithFields(logrus.Fields{
				"collector": name,
			}).Info("collector config not found")
		}
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
