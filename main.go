package main

import (
	"fmt"
	"os"
	"time"

	"github.com/GannettDigital/go-newrelic-plugin/collectors"
	"github.com/Sirupsen/logrus"
	newrelicMonitoring "github.com/newrelic/go-agent"
	"github.com/spf13/viper"
)

var log = logrus.New()

func main() {

	config := loadConfig()

	app := setupNewRelic(config)

	// main routine
	for name, collector := range collectors.CollectorArray {
		// if config file has enabled the collector indicated by collectorName
		if config.Collectors[name].Enabled {
			go func(collectorName string, collectorValue collectors.Collector) {
				// TODO: random delay to offset collections
				ticker := time.NewTicker(readCollectorDelay(collectorName, config))
				for _ = range ticker.C {
					go getResult(collectorName, app, config, collectorValue)
				}
			}(name, collector) // you must close over this variable or it will change on the function when the next iteration occurs https://github.com/golang/go/wiki/CommonMistakes
		}
	}

	done := make(chan bool)
	<-done // block forever

}

func readCollectorDelay(name string, conf collectors.Config) time.Duration {
	collectorConf := conf.Collectors[name]
	delay := conf.DefaultDelayMS

	if collectorConf.DelayMS != 0 {
		delay = collectorConf.DelayMS
	}

	result, err := time.ParseDuration(fmt.Sprintf("%dms", delay))

	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("Delay time parsing error")
	}
	return result
}

func getResult(collectorName string, app newrelicMonitoring.Application, config collectors.Config, collector collectors.Collector) {
	c := make(chan []map[string]interface{}, 1)
	collector(config, c)

	select {
	case responses, success := <-c:
		if success {
			log.WithFields(logrus.Fields{
				"collector": collectorName,
			}).Info("received data from collector")
			for _, response := range responses {
				sendData(collectorName, app, config, response)
			}
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
	cfg := newrelicMonitoring.NewConfig(config.AppName, config.NewRelicKey)

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

		os.Exit(1)
	}

	return app
}

// loadConfig - read from config file and marshal info collectors.Config
func loadConfig() collectors.Config {
	// set up viper to find config file
	vip := viper.New()
	vip.SetConfigType("yaml")
	vip.SetConfigName("config")
	vip.AddConfigPath("/etc/newrelic_plugins/")
	vip.AddConfigPath(".")

	// read in config file
	err := vip.ReadInConfig()
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("Error reading config file")

		os.Exit(1)
	}

	// marshal config file data into collectors.Config and return it
	var conf collectors.Config
	err = vip.Unmarshal(&conf)
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("Error unmarshaling configs")

		os.Exit(1)
	}
	return conf
}
