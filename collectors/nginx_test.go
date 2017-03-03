package collectors

import (
	"reflect"
	"testing"

	fake "github.com/GannettDigital/go-newrelic-plugin/collectors/fake"
	"github.com/franela/goblin"
)

var fakeFullConfig Config
var fakeConfig NginxConfig

func init() {
	fakeConfig = NginxConfig{
		NginxListenPort: "8140",
		NginxStatusURI:  "nginx_status",
		NginxStatusPage: "http://localhost",
	}
	fakeFullConfig = Config{
		AppName:        "test-newrelic-plugin",
		NewRelicKey:    "somenewrelickeyhere",
		DefaultDelayMS: 1000,
		Collectors: map[string]CommonConfig{
			"nginx": CommonConfig{
				Enabled:         true,
				DelayMS:         500,
				CollectorConfig: fakeConfig,
			},
		},
	}
}

func TestNginxCollector(t *testing.T) {
	g := goblin.Goblin(t)

	resultSlice := make([]map[string]interface{}, 1)
	resultSlice[0] = map[string]interface{}{
		"nginx.net.connections": 2,
		"nginx.net.accepts":     29,
		"nginx.net.handled":     29,
		"nginx.net.requests":    31,
		"nginx.net.writing":     1,
		"nginx.net.waiting":     1,
		"nginx.net.reading":     0,
	}

	var tests = []struct {
		HTTPRunner      fake.HTTPResult
		ExpectedResult  []map[string]interface{}
		TestDescription string
	}{
		{
			HTTPRunner: fake.HTTPResult{
				Code: 200,
				Data: []byte("Active connections: 2 \nserver accepts handled requests\n 29 29 31 \nReading: 0 Writing: 1 Waiting: 1 "),
			},
			ExpectedResult:  resultSlice,
			TestDescription: "Successfully GET Nginx status page",
		},
	}

	for _, test := range tests {
		g.Describe("NginxCollector()", func() {
			g.It(test.TestDescription, func() {
				runner = test.HTTPRunner
				stats := make(chan []map[string]interface{}, 1)
				NginxCollector(fakeFullConfig, stats)
				close(stats)
				for stat := range stats {
					g.Assert(reflect.DeepEqual(stat, test.ExpectedResult)).Equal(true)
				}
			})
		})
	}
}

func TestGetNginxStatus(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		HTTPRunner      fake.HTTPResult
		TestDescription string
	}{
		{
			HTTPRunner: fake.HTTPResult{
				Code: 200,
				Data: []byte("Active connections: 2 \nserver accepts handled requests\n 29 29 31 \nReading: 0 Writing: 1 Waiting: 1 "),
			},
			TestDescription: "Successfully GET Nginx status page",
		},
	}

	for _, test := range tests {
		g.Describe("getNginxStatus()", func() {
			g.It(test.TestDescription, func() {
				runner = test.HTTPRunner
				result := getNginxStatus(fakeConfig, make(chan []map[string]interface{}, 1))
				g.Assert(reflect.DeepEqual(result, string(test.HTTPRunner.Data))).Equal(true)
			})
		})
	}
}

func TestScrapeStatus(t *testing.T) {
	g := goblin.Goblin(t)

	resultSlice := make([]map[string]interface{}, 1)
	resultSlice[0] = map[string]interface{}{
		"nginx.net.connections": 2,
		"nginx.net.accepts":     29,
		"nginx.net.handled":     29,
		"nginx.net.requests":    31,
		"nginx.net.writing":     1,
		"nginx.net.waiting":     1,
		"nginx.net.reading":     0,
	}

	var tests = []struct {
		Data            string
		ExpectedResult  []map[string]interface{}
		TestDescription string
	}{
		{
			Data:            "Active connections: 2 \nserver accepts handled requests\n 29 29 31 \nReading: 0 Writing: 1 Waiting: 1 ",
			ExpectedResult:  resultSlice,
			TestDescription: "Successfully scrape given status page",
		},
	}

	for _, test := range tests {
		g.Describe("scrapeStatus()", func() {
			g.It(test.TestDescription, func() {
				result := scrapeStatus(test.Data)
				g.Assert(reflect.DeepEqual(result, test.ExpectedResult)).Equal(true)
			})
		})
	}
}

func TestToInt(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		Value           string
		ExpectedResult  int
		TestDescription string
	}{
		{
			Value:           "234567",
			ExpectedResult:  234567,
			TestDescription: "Should return int 234567 of string",
		},
		{
			Value:           "",
			ExpectedResult:  0,
			TestDescription: "Should return 0 if empty string",
		},
		{
			Value:           "xyz",
			ExpectedResult:  0,
			TestDescription: "Should return 0 if error converting to int",
		},
	}

	for _, test := range tests {
		g.Describe("toInt()", func() {
			g.It(test.TestDescription, func() {
				result := toInt(test.Value)
				g.Assert(reflect.DeepEqual(result, test.ExpectedResult)).Equal(true)
			})
		})
	}
}
