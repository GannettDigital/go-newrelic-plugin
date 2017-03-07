package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/GannettDigital/go-newrelic-plugin/collectors"
	fakeGofigure "github.com/GannettDigital/goFigure/fake"
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
		InputName       string
		InputClient     *fakeGofigure.ConfigClient
		InputBucket     string
		InputItemPath   string
		ExpectedConfig  collectors.Config
		ExpectedErr     error
		TestDescription string
	}{
		{
			InputName:     "",
			InputClient:   &fakeGofigure.ConfigClient{},
			InputBucket:   "",
			InputItemPath: "",
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
					"couchbase": collectors.CommonConfig{
						Enabled: true,
						DelayMS: 30000,
						CollectorConfig: map[string]interface{}{
							"couchbaseuser":     "admin",
							"couchbasepassword": "password",
							"couchbaseport":     "8091",
							"couchbasehost":     "http://10.84.103.211"},
					},
					"rabbitmq": collectors.CommonConfig{
						Enabled: false,
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
			ExpectedErr:     nil,
			TestDescription: "Should successfully load a config from file",
		},
		{
			InputName:       "somenoexistyfile",
			InputClient:     &fakeGofigure.ConfigClient{},
			InputBucket:     "somebucket",
			InputItemPath:   "somepath/config.yaml",
			ExpectedConfig:  collectors.Config{},
			ExpectedErr:     nil,
			TestDescription: "Should successfully load and unmarshall a config file",
		},
		{
			InputName: "somenoexistyfile",
			InputClient: &fakeGofigure.ConfigClient{
				Err: errors.New("some error"),
			},
			InputBucket:     "somebucket",
			InputItemPath:   "somepath/config.yaml",
			ExpectedConfig:  collectors.Config{},
			ExpectedErr:     errors.New("some error"),
			TestDescription: "Should return an error when one is encountered from goFigure",
		},
		{
			InputName:       "somenoexistyfile",
			InputClient:     &fakeGofigure.ConfigClient{},
			InputBucket:     "",
			InputItemPath:   "",
			ExpectedConfig:  collectors.Config{},
			ExpectedErr:     errors.New("No configs located"),
			TestDescription: "Should fail to find configs when no config with the given name is found",
		},
	}

	for _, test := range tests {
		g.Describe("loadConfig()", func() {
			g.It(test.TestDescription, func() {
				conf, err := loadConfig(test.InputName, test.InputClient, test.InputBucket, test.InputItemPath)
				g.Assert(conf).Equal(test.ExpectedConfig)
				g.Assert(err).Equal(test.ExpectedErr)
			})
		})
	}
}

func TestProcessTags(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputName       string
		InputConfig     collectors.Config
		ExpectedRes     map[string]interface{}
		TestDescription string
	}{
		{
			InputName: "somecollector",
			InputConfig: collectors.Config{
				Tags: collectors.Tags{
					KeyValue: map[string]string{
						"kvglobal1": "kvvalue1",
						"kvglobal2": "kvvalue2",
					},
					Env: []string{
						"GLOBAL_1",
					},
				},
				Collectors: map[string]collectors.CommonConfig{
					"somecollector": collectors.CommonConfig{
						Tags: collectors.Tags{
							KeyValue: map[string]string{
								"kvlocal1": "kvlocalvalue1",
							},
							Env: []string{
								"LOCAL_1",
								"LOCAL_2",
							},
						},
					},
				},
			},
			ExpectedRes: map[string]interface{}{
				"global_1":  "global_1value",
				"local_1":   "local_1value",
				"local_2":   "local_2value",
				"kvglobal2": "kvvalue2",
				"kvglobal1": "kvvalue1",
				"kvlocal1":  "kvlocalvalue1",
			},
			TestDescription: "Should successfully process all kv and env tags for both global and collector scope",
		},
	}

	for _, test := range tests {
		g.Describe("processTags()", func() {
			g.It(test.TestDescription, func() {
				fullList := append(test.InputConfig.Tags.Env, test.InputConfig.Collectors[test.InputName].Tags.Env...)
				for _, env := range fullList {
					err := os.Setenv(env, fmt.Sprintf("%svalue", strings.ToLower(env)))

					g.Assert(err).Equal(nil)
				}

				res := processTags(test.InputName, test.InputConfig)
				g.Assert(res).Equal(test.ExpectedRes)

				for _, env := range fullList {
					err := os.Unsetenv(env)

					g.Assert(err).Equal(nil)
				}
			})
		})
	}
}

func TestMergeMaps(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputMap1       map[string]interface{}
		InputMap2       map[string]interface{}
		ExpectedRes     map[string]interface{}
		TestDescription string
	}{
		{
			InputMap1: map[string]interface{}{
				"thing1": "stuff",
			},
			InputMap2: map[string]interface{}{
				"thing2": "otherstuff",
			},
			ExpectedRes: map[string]interface{}{
				"thing1": "stuff",
				"thing2": "otherstuff",
			},
			TestDescription: "Should successfully merge 2 maps",
		},
	}

	for _, test := range tests {
		g.Describe("mergeMaps()", func() {
			g.It(test.TestDescription, func() {
				res := mergeMaps(test.InputMap1, test.InputMap2)
				g.Assert(res).Equal(test.ExpectedRes)
			})
		})
	}
}

func TestReadEnvList(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputList       []string
		ExpectedRes     map[string]interface{}
		TestDescription string
	}{
		{
			InputList: []string{
				"ENV_1",
				"ENV_2",
			},
			ExpectedRes: map[string]interface{}{
				"env_1": "env_1value",
				"env_2": "env_2value",
			},
			TestDescription: "Should successfully format and return properly formatted and read env values",
		},
	}

	for _, test := range tests {
		g.Describe("readEnvList()", func() {
			g.It(test.TestDescription, func() {
				for _, env := range test.InputList {
					err := os.Setenv(env, fmt.Sprintf("%svalue", strings.ToLower(env)))

					g.Assert(err).Equal(nil)
				}

				res := readEnvList(test.InputList)
				g.Assert(res).Equal(test.ExpectedRes)

				for _, env := range test.InputList {
					err := os.Unsetenv(env)

					g.Assert(err).Equal(nil)
				}
			})
		})
	}
}

func TestConvertToInterfaceMap(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputMap        map[string]string
		ExpectedRes     map[string]interface{}
		TestDescription string
	}{
		{
			InputMap: map[string]string{
				"thing1": "stuff",
				"thing2": "morestuff",
			},
			ExpectedRes: map[string]interface{}{
				"thing1": "stuff",
				"thing2": "morestuff",
			},
			TestDescription: "Should successfully perform a conversion to interfaces",
		},
	}

	for _, test := range tests {
		g.Describe("convertToInterfaceMap()", func() {
			g.It(test.TestDescription, func() {
				res := convertToInterfaceMap(test.InputMap)
				g.Assert(res).Equal(test.ExpectedRes)
			})
		})
	}
}
