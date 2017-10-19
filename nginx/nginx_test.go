package nginx

import (
	"fmt"
	"reflect"
	"testing"

	fake "github.com/GannettDigital/paas-api-utils/utilsHTTP/fake"
	"github.com/Sirupsen/logrus"
	"github.com/franela/goblin"
)

var fakeConfig Config

func init() {
	fakeConfig = Config{
		NginxListenPort: "8140",
		NginxStatusURI:  "nginx_status",
		NginxHost:       "http://localhost",
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
				ResultsList: []fake.Result{
					{
						Method: "GET",
						URI:    "/nginx_status",
						Code:   200,
						Data:   []byte("Active connections: 2 \nserver accepts handled requests\n 29 29 31 \nReading: 0 Writing: 1 Waiting: 1 "),
						Err:    nil,
					},
				},
			},
			TestDescription: "Successfully GET Nginx status page",
		},
	}

	for _, test := range tests {
		g.Describe("getNginxStatus()", func() {
			g.It(test.TestDescription, func() {
				runner = test.HTTPRunner
				result := getNginxStatus(logrus.New(), fakeConfig)
				g.Assert(reflect.DeepEqual(result, string(test.HTTPRunner.ResultsList[0].Data))).Equal(true)
			})
		})
	}
}

func TestScrapeStatus(t *testing.T) {
	g := goblin.Goblin(t)

	result := map[string]interface{}{
		"event_type":            "LoadBalancerSample",
		"provider":              "nginx",
		"nginx.hostname":        "",
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
		ExpectedResult  map[string]interface{}
		TestDescription string
	}{
		{
			Data:            "Active connections: 2 \nserver accepts handled requests\n 29 29 31 \nReading: 0 Writing: 1 Waiting: 1 ",
			ExpectedResult:  result,
			TestDescription: "Successfully scrape given status page",
		},
	}

	for _, test := range tests {
		g.Describe("scrapeStatus()", func() {
			g.It(test.TestDescription, func() {
				result := scrapeStatus(logrus.New(), test.Data)
				fmt.Println(result)
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
				result := toInt(logrus.New(), test.Value)
				g.Assert(reflect.DeepEqual(result, test.ExpectedResult)).Equal(true)
			})
		})
	}
}
