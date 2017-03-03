package main

import (
	"testing"
	"time"

	"github.com/GannettDigital/go-newrelic-plugin/collectors"
	"github.com/franela/goblin"
)

func TestReadCollectorDelay(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputName        string
		InputConfig      collectors.Config
		ExpectedDuration time.Duration
		TestDescription  string
	}{
		{
			InputName: "testy",
			InputConfig: collectors.Config{
				AppName:        "test-newrelic-plugin",
				NewRelicKey:    "somenewrelickeyhere",
				DefaultDelayMS: 1000,
				Collectors: map[string]collectors.CommonConfig{
					"testy": collectors.CommonConfig{
						Enabled: true,
						CollectorConfig: map[string]interface{}{
							"otherthing": "thing",
							"something":  "stuff"},
					},
				},
			},
			ExpectedDuration: (time.Millisecond * 1000),
			TestDescription:  "Should successfully return the expected duration when no duration overide is set",
		},
		{
			InputName: "testy",
			InputConfig: collectors.Config{
				AppName:        "test-newrelic-plugin",
				NewRelicKey:    "somenewrelickeyhere",
				DefaultDelayMS: 1000,
				Collectors: map[string]collectors.CommonConfig{
					"testy": collectors.CommonConfig{
						Enabled: true,
						DelayMS: 500,
						CollectorConfig: map[string]interface{}{
							"otherthing": "thing",
							"something":  "stuff"},
					},
				},
			},
			ExpectedDuration: (time.Millisecond * 500),
			TestDescription:  "Should successfully return the expected duration when an override value exists",
		},
	}

	for _, test := range tests {
		g.Describe("readCollectorDelay()", func() {
			g.It(test.TestDescription, func() {
				res := readCollectorDelay(test.InputName, test.InputConfig)
				g.Assert(res).Equal(test.ExpectedDuration)
			})
		})
	}
}

func TestLoadConfig(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		ExpectedConfig  collectors.Config
		TestDescription string
	}{
		{
			ExpectedConfig: collectors.Config{
				AppName:        "test-newrelic-plugin",
				NewRelicKey:    "somenewrelickeyhere",
				DefaultDelayMS: 1000,
				Tags: collectors.Tags{
					KeyValue: map[string]string{
						"tag2": "someothertagvalue",
						"tag1": "sometagvalue",
					},
					Env: []string{
						"VAR_1",
						"VAR_2"},
				},
				Collectors: map[string]collectors.CommonConfig{
					"haproxy": collectors.CommonConfig{
						Enabled: false,
						DelayMS: 1000,
						Tags: collectors.Tags{
							KeyValue: map[string]string{
								"tag4": "someothertagvalue4",
								"tag3": "sometagvalue3",
							},
							Env: []string{
								"VAR_3", "VAR_4",
							},
						},
						CollectorConfig: map[string]interface{}{
							"otherthing": "thing",
							"something":  "stuff"},
					},
					"rabbitmq": collectors.CommonConfig{
						Enabled: true,
						DelayMS: 2000,
						CollectorConfig: map[string]interface{}{
							"rabbitmquser":     "scalr",
							"rabbitmqpassword": "hiTVPamzPm",
							"rabbitmqport":     "15672",
							"rabbitmqhost":     "http://10.84.100.59"},
					},
					"nginx": collectors.CommonConfig{
						Enabled: false,
						DelayMS: 1000,
						CollectorConfig: map[string]interface{}{
							"nginxlistenport": "8140",
							"nginxstatusuri":  "nginx_status",
							"nginxstatuspage": "http://localhost"},
					},
				},
			},
			TestDescription: "Should successfully load a config from file",
		},
	}

	for _, test := range tests {
		g.Describe("loadConfig()", func() {
			g.It(test.TestDescription, func() {
				conf := loadConfig()
				g.Assert(conf).Equal(test.ExpectedConfig)
			})
		})
	}
}
