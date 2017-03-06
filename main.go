package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/GannettDigital/go-newrelic-plugin/collectors"
	"github.com/GannettDigital/goFigure"
	"github.com/Sirupsen/logrus"
	newrelicMonitoring "github.com/newrelic/go-agent"
	"github.com/spf13/viper"
)

var log = logrus.New()

func main() {

	config, err := loadConfig("", &gofigure.ConfigClient{}, os.Getenv("GOFIGURE_BUCKET"), os.Getenv("GOFIGURE_ITEM_PATH"))
	if err != nil {
		os.Exit(1)
	}

	app := setupNewRelic(config)

	// main routine
	for name, collector := range collectors.CollectorArray {
		// if config file has enabled the collector indicated by collectorName
		if config.Collectors[name].Enabled {
			go func(collectorName string, collectorValue collectors.Collector) {
				rand.Seed(time.Now().UnixNano())
				sleepTime := time.Millisecond * time.Duration(rand.Intn(1000))
				log.Info(fmt.Sprintf("sleeping %s to offset collector collections", sleepTime))
				time.Sleep(sleepTime)
				ticker := time.NewTicker(readCollectorDelay(collectorName, config))
				for _ = range ticker.C {
					getResult(collectorName, app, config, collectorValue)
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

	defer func() {
		// recover from panic if one occured. Set err to nil otherwise.
		if err := recover(); err != nil {
			log.WithFields(logrus.Fields{
				"error": err,
			}).Error("collector panic'd, bad collector..")
			// if channel was closed, it will panic, so just collect it and go on our way
			defer func() {
				if err := recover(); err != nil {
					log.WithFields(logrus.Fields{
						"error": err,
					}).Info("collector already closed channel")
				}
			}()
			close(c)
		}
	}()

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

	tags := processTags(collectorName, config)
	payload := mergeMaps(tags, stats)
	// send stats
	app.RecordCustomEvent(fmt.Sprintf("gannettNewRelic%s", collectorName), payload)
}

// processTags - read all ENV tags, and append collector specific to global for both kv tags and env tags
func processTags(collectorName string, config collectors.Config) map[string]interface{} {
	kvList := mergeMaps(convertToInterfaceMap(config.Tags.KeyValue), convertToInterfaceMap(config.Collectors[collectorName].Tags.KeyValue))
	envList := append(config.Tags.Env, config.Collectors[collectorName].Tags.Env...)
	kvList = mergeMaps(kvList, readEnvList(envList))
	return kvList
}

// mergeMaps - merge two map[string]interface{}
func mergeMaps(global map[string]interface{}, specific map[string]interface{}) map[string]interface{} {
	for key, value := range specific {
		global[key] = value
	}
	return global
}

// readEnvList - read all environment variables and return them as a map[string]interface{}
func readEnvList(envList []string) map[string]interface{} {
	resultList := make(map[string]interface{})
	for _, env := range envList {
		resultList[strings.ToLower(env)] = os.Getenv(env)
	}
	return resultList
}

// convertToInterfaceMap - make map[string]string into a map[string]interface
func convertToInterfaceMap(stringMap map[string]string) map[string]interface{} {
	interfaceMap := make(map[string]interface{})
	for key, value := range stringMap {
		interfaceMap[key] = value
	}
	return interfaceMap
}

func setupNewRelic(config collectors.Config) newrelicMonitoring.Application {

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
func loadConfig(configName string, client gofigure.Client, bucket string, itemPath string) (collectors.Config, error) {
	if configName == "" {
		configName = "config"
	}
	// config object to return
	var conf collectors.Config

	// set up viper to find config file
	vip := viper.New()
	vip.SetConfigType("yaml")
	vip.SetConfigName(configName)
	vip.AddConfigPath("/etc/newrelic_plugins/")
	vip.AddConfigPath(".")

	// read in config file
	err := vip.ReadInConfig()
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Info("Error reading local config file, will attempt to locate goFigure configs...")

		// attempt to read from s3 bucket if GOFIGURE_BUCKET and GOFIGURE_ITEM_PATH are set
		if (bucket != "") && (itemPath != "") {
			client.Setup(bucket, itemPath, "yaml")

			err = client.LoadAndUnmarshal(&conf)
			if err != nil {
				log.WithFields(logrus.Fields{
					"err": err,
				}).Error("Error loading config file from s3")

				return conf, err
			}

			return conf, nil
		}

		log.WithFields(logrus.Fields{}).Error("Error locating any configs")
		return conf, errors.New("No configs located")
	}

	// marshal config file data into collectors.Config and return it
	err = vip.Unmarshal(&conf)
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("Error unmarshaling configs")

		return conf, err
	}
	return conf, nil
}
