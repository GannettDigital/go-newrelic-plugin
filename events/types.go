package events

import (
	"github.com/GannettDigital/go-newrelic-plugin/helpers"
	newrelicMonitoring "github.com/newrelic/go-agent"
)

type Config struct {
	App         newrelicMonitoring.Application
	AppName     string
	NginxConfig helpers.NginxConfig
}
