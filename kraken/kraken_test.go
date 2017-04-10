package kraken

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
		KrakenListenPort: "8140",
		KrakenHost:       "http://localhost",
	}
}

func TestGetKrakenStatus(t *testing.T) {
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
						URI:    "/",
						Code:   200,
						Data:   []byte("Load Test Started: 25.0 seconds ago\n\nVersion: 2.2.0\nCustomer: None\nProject: None\nState: Running\nTest duration: 0:00:25\nSamples count: 178, 100.00% failures\nAverage times: total 0.106, latency 0.106, connect 0.000\nPercentile 0.0%: 0.037\nPercentile 50.0%: 0.120\nPercentile 90.0%: 0.125\nPercentile 95.0%: 0.126\nPercentile 99.0%: 0.167\nPercentile 99.9%: 0.281\nPercentile 100.0%: 0.281 "),
						Err:    nil,
					},
				},
			},
			TestDescription: "Successfully GET kraken status page",
		},
	}

	for _, test := range tests {
		g.Describe("getKrakenStatus()", func() {
			g.It(test.TestDescription, func() {
				runner = test.HTTPRunner
				result := getKrakenStatus(logrus.New(), fakeConfig)
				g.Assert(reflect.DeepEqual(result, string(test.HTTPRunner.ResultsList[0].Data))).Equal(true)
			})
		})
	}
}

func TestScrapeStatus(t *testing.T) {
	g := goblin.Goblin(t)

	result := map[string]interface{}{
		"event_type":            				"GKrakenSample",
		"provider":              				"kraken",
		"kraken.version":						    "2.2.0",
		"kraken.customer":			   	    "None",
		"kraken.project":						    "None",
		"kraken.state":				   		    "Running",
		"kraken.kpi.avg_resp_time":     0.106,
		"kraken.kpi.avg_latency":       0.106,
		"kraken.kpi.avg_conn_time":     0.000,
		"kraken.kpi.percentiles.50":    0.120,
		"kraken.kpi.percentiles.90":    0.125,
		"kraken.kpi.percentiles.95":    0.126,
		"kraken.kpi.percentiles.99":    0.167,
		"kraken.kpi.percentiles.100":   0.281,
		"kraken.sample_count":          178,
		"kraken.sample_failure":        100.00,
		"kraken.duration":              "0:00:25",
	}

	var tests = []struct {
		Data            string
		ExpectedResult  map[string]interface{}
		TestDescription string
	}{
		{
			Data:            "Load Test Started: 25.0 seconds ago\n\nVersion: 2.2.0\nCustomer: None\nProject: None\nState: Running\nTest duration: 0:00:25\nSamples count: 178, 100.00. failures\nAverage times: total 0.106, latency 0.106, connect 0.000\nPercentile 0.0%: 0.037\nPercentile 50.0%: 0.120\nPercentile 90.0%: 0.125\nPercentile 95.0%: 0.126\nPercentile 99.0%: 0.167\nPercentile 99.9%: 0.281\nPercentile 100.0%: 0.281 ",
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
