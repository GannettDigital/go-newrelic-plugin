package metrics

import (
	"reflect"
	"testing"

	"github.com/GannettDigital/go-newrelic-plugin/helpers"
	fake "github.com/GannettDigital/go-newrelic-plugin/metrics/fake"
	"github.com/franela/goblin"
)

var fakeConfig helpers.NGINXConfig

func init() {
	fakeConfig = helpers.NGINXConfig{
		NGINXListenPort: "8140",
		NGINXStatusURI:  "nginx_status",
		NGINXStatusPage: "http://localhost",
	}
}

func TestPollStatus(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		HTTPRunner      fake.HTTPResult
		ExpectedResult  helpers.NginxMetrics
		TestDescription string
	}{
		{
			HTTPRunner: fake.HTTPResult{
				Code: 200,
				Data: []byte("Active connections: 2 \nserver accepts handled requests\n 29 29 31 \nReading: 0 Writing: 1 Waiting: 1 "),
			},
			ExpectedResult: helpers.NginxMetrics{
				Connections: 2,
				Accepts:     29,
				Handled:     29,
				Requests:    31,
				Writing:     1,
				Waiting:     1,
				Reading:     0,
			},
			TestDescription: "Successfully GET NGINX status page",
		},
	}

	for _, test := range tests {
		g.Describe("PollStatus()", func() {
			g.It(test.TestDescription, func() {
				result := PollStatus(&fakeConfig, test.HTTPRunner)
				g.Assert(reflect.DeepEqual(result, test.ExpectedResult)).Equal(true)
			})
		})
	}
}

func TestGetNGINXStatus(t *testing.T) {
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
			TestDescription: "Successfully GET NGINX status page",
		},
	}

	for _, test := range tests {
		g.Describe("getNGINXStatus()", func() {
			g.It(test.TestDescription, func() {
				result := getNGINXStatus(&fakeConfig, test.HTTPRunner)
				g.Assert(reflect.DeepEqual(result, string(test.HTTPRunner.Data))).Equal(true)
			})
		})
	}
}

func TestScrapeStatus(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		Data            string
		ExpectedResult  helpers.NginxMetrics
		TestDescription string
	}{
		{
			Data: "Active connections: 2 \nserver accepts handled requests\n 29 29 31 \nReading: 0 Writing: 1 Waiting: 1 ",
			ExpectedResult: helpers.NginxMetrics{
				Connections: 2,
				Accepts:     29,
				Handled:     29,
				Requests:    31,
				Writing:     1,
				Waiting:     1,
				Reading:     0,
			},
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
