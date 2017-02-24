package main

import (
	"fmt"
	"os"
	"time"

	"github.com/GannettDigital/go-newrelic-plugin/events"
	"github.com/GannettDigital/go-newrelic-plugin/helpers"
	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	newrelicMonitoring "github.com/newrelic/go-agent"
)

/*
NOTES:
In the future stuff like App Name, Nginx settings should probably be pulled from config.yaml
Default app name should actually make sense (app, env, etc)
*/

var log = logrus.New()

const (
	// App name
	DefaultAppName = "go-newrelic-plugin"
	// NewRelic settings
	DefaultNewRelicKey = "FAKELICENSEKEYFAKELICENSEKEYFAKELICENSEK"
	// Nginx settings
	DefaultNGINXListenPort = "8140"
	DefaultNGINXStatusURI  = "nginx_status"
	DefaultNGINXStatusPage = "http://localhost"
)

// Options -
var opts struct {
	// optional
	AppName         string `long:"app-name"`
	NewRelicKey     string `long:"new-relic-key"`
	NGINXListenPort string `long:"nginx-listen-port"`
	NGINXStatusURI  string `long:"nginx-status-uri"`
	NGINXStatusPage string `long:"nginx-status-page"`
}

func init() {
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}
	if opts.AppName == "" {
		opts.AppName = DefaultAppName
	}
	if opts.NewRelicKey == "" {
		opts.NewRelicKey = DefaultNewRelicKey
	}
	if opts.NGINXListenPort == "" {
		opts.NGINXListenPort = DefaultNGINXListenPort
	}
	if opts.NGINXStatusURI == "" {
		opts.NGINXStatusURI = DefaultNGINXStatusURI
	}
	if opts.NGINXStatusPage == "" {
		opts.NGINXStatusPage = DefaultNGINXStatusPage
	}
}

func getMetrics(app newrelicMonitoring.Application) {
	pluginConfig := events.Config{
		App:     app,
		AppName: opts.AppName,
		NGINXConfig: helpers.NGINXConfig{
			NGINXListenPort: opts.NGINXListenPort,
			NGINXStatusURI:  opts.NGINXStatusURI,
			NGINXStatusPage: opts.NGINXStatusPage,
		},
	}

	for {
		pluginConfig.NGINXEvent()
		time.Sleep(time.Minute)
	}
}

func main() {
	// Create an app config.  Application name and New Relic license key are required.
	cfg := newrelicMonitoring.NewConfig(opts.AppName, opts.NewRelicKey)

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
		fmt.Println(err)
		os.Exit(1)
	}

	if err := app.WaitForConnection(10 * time.Second); nil != err {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("Connection error")
	}

	getMetrics(app)
}
