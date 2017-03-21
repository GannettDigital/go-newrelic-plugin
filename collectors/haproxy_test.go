package collectors

import (
	"reflect"
	"testing"

	fake "github.com/GannettDigital/go-newrelic-plugin/collectors/fake"
	"github.com/franela/goblin"
)

var fakeFullConfig Config
var fakeConfig HaproxyConfig

func init() {
	fakeConfig = HaproxyConfig{
		HaproxyStatsPort:  "8000",
		HaproxyStatusURI:  "haproxy;csv",
		HaproxyStatusPage: "http://localhost",
	}
	fakeFullConfig = Config{
		AppName:        "test-newrelic-plugin",
		NewRelicKey:    "somenewrelickeyhere",
		DefaultDelayMS: 1000,
		Collectors: map[string]CommonConfig{
			"haproxy": CommonConfig{
				Enabled:         true,
				DelayMS:         500,
				CollectorConfig: fakeConfig,
			},
		},
	}
}

func TestHaproxyCollector(t *testing.T) {
	g := goblin.Goblin(t)

	resultSlice := make([]map[string]interface{}, 1)
	resultSlice[0] = map[string]interface{}{
		"haproxy.frontend.session.current":    30,
		"haproxy.frontend.session.max":        134,
		"haproxy.frontend.session.limit":      50000,
		"haproxy.frontend.session.total":      18385,
		"haproxy.frontend.bytes.in_rate":      602550387,
		"haproxy.frontend.bytes.out_rate":     4248437204,
		"haproxy.frontend.denied.req_rate":    0,
		"haproxy.frontend.denied.resp_rate":   0,
		"haproxy.frontend.errors.req_rate":    35,
		"haproxy.frontend.session.rate":       0,
		"haproxy.frontend.response.1xx":       0,
		"haproxy.frontend.response.2xx":       124940,
		"haproxy.frontend.response.3xx":       144833,
		"haproxy.frontend.response.4xx":       9148,
		"haproxy.frontend.response.5xx":       239,
		"haproxy.frontend.response.other":     0,
		"haproxy.frontend.requests.rate":      15,
		"haproxy.backend.queue.current":       0,
		"haproxy.backend.queue.max":           0,
		"haproxy.backend.session.current":     5,
		"haproxy.backend.session.max":         73,
		"haproxy.backend.session.limit":       5000,
		"haproxy.backend.session.total":       279131,
		"haproxy.backend.bytes.in_rate":       602498165,
		"haproxy.backend.bytes.out_rate":      4248430659,
		"haproxy.backend.denied.req_rate":     0,
		"haproxy.backend.denied.resp_rate":    0,
		"haproxy.backend.errors.con_rate":     0,
		"haproxy.backend.errors.resp_rate":    0,
		"haproxy.backend.warnings.retr_rate":  0,
		"haproxy.backend.warnings.redis_rate": 0,
		"haproxy.backend.session.rate":        15,
		"haproxy.backend.response.1xx":        0,
		"haproxy.backend.response.2xx":        124940,
		"haproxy.backend.response.3xx":        144833,
		"haproxy.backend.response.4xx":        9113,
		"haproxy.backend.response.5xx":        239,
		"haproxy.backend.response.other":      0,
		"haproxy.backend.queue.time":          0,
		"haproxy.backend.connect.time":        1,
		"haproxy.backend.response.time":       372,
		"haproxy.backend.session.time":        1343,
	}

	var tests = []struct {
		HTTPRunner      fake.HTTPResult
		ExpectedResult  []map[string]interface{}
		TestDescription string
	}{
		{
			HTTPRunner: fake.HTTPResult{
				Code: 200,
				Data: []byte("# pxname,svname,qcur,qmax,scur,smax,slim,stot,bin,bout,dreq,dresp,ereq,econ,eresp,wretr,wredis,status,weight,act,bck,chkfail,chkdown,lastchg,downtime,qlimit,pid,iid,sid,throttle,lbtot,tracked,type,rate,rate_lim,rate_max,check_status,check_code,check_duration,hrsp_1xx,hrsp_2xx,hrsp_3xx,hrsp_4xx,hrsp_5xx,hrsp_other,hanafail,req_rate,req_rate_max,req_tot,cli_abrt,srv_abrt,comp_in,comp_out,comp_byp,comp_rsp,lastsess,last_chk,last_agt,qtime,ctime,rtime,ttime,\n
					stats,FRONTEND,,,1,1,50000,13938,3693800,17784897,0,0,0,,,,,OPEN,,,,,,,,,1,2,0,,,,0,1,0,4,,,,0,13933,0,0,5,0,,1,4,13939,,,0,0,0,0,,,,,,,,\n
					stats,BACKEND,0,0,0,1,5000,5,3693800,17784897,0,0,,5,0,0,0,UP,0,0,0,,0,266274,0,,1,2,0,,0,,1,0,,3,,,,0,0,0,0,5,0,,,,,0,0,0,0,0,0,0,,,0,0,0,0,\n
					http_frontend,FRONTEND,,,0,0,50000,0,0,0,0,0,0,,,,,OPEN,,,,,,,,,1,3,0,,,,0,0,0,0,,,,0,0,0,0,0,0,,0,0,0,,,0,0,0,0,,,,,,,,\n
					unsecure,member:2:10.84.91.161,0,0,0,0,15000,0,0,0,,0,,0,0,0,0,UP,1,1,0,0,0,266274,0,,1,4,1,,0,,2,0,,0,L7OK,200,2,0,0,0,0,0,0,0,,,,0,0,,,,,-1,OK,,0,0,0,0,\n
					unsecure,BACKEND,0,0,0,0,5000,0,0,0,0,0,,0,0,0,0,UP,1,1,0,,0,266274,0,,1,4,0,,0,,1,0,,0,,,,0,0,0,0,0,0,,,,,0,0,0,0,0,0,-1,,,0,0,0,0,"),
			},
			ExpectedResult:  resultSlice,
			TestDescription: "Successfully GET Haproxy status page",
		},
	}

	for _, test := range tests {
		g.Describe("HaproxyCollector()", func() {
			g.It(test.TestDescription, func() {
				runner = test.HTTPRunner
				stats := make(chan []map[string]interface{}, 1)
				HaproxyCollector(fakeFullConfig, stats)
				close(stats)
				for stat := range stats {
					g.Assert(reflect.DeepEqual(stat, test.ExpectedResult)).Equal(true)
				}
			})
		})
	}
}

func TestGetHaproxyStatus(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		HTTPRunner      fake.HTTPResult
		TestDescription string
	}{
		{
			HTTPRunner: fake.HTTPResult{
				Code: 200,
				Data: []byte("# pxname,svname,qcur,qmax,scur,smax,slim,stot,bin,bout,dreq,dresp,ereq,econ,eresp,wretr,wredis,status,weight,act,bck,chkfail,chkdown,lastchg,downtime,qlimit,pid,iid,sid,throttle,lbtot,tracked,type,rate,rate_lim,rate_max,check_status,check_code,check_duration,hrsp_1xx,hrsp_2xx,hrsp_3xx,hrsp_4xx,hrsp_5xx,hrsp_other,hanafail,req_rate,req_rate_max,req_tot,cli_abrt,srv_abrt,comp_in,comp_out,comp_byp,comp_rsp,lastsess,last_chk,last_agt,qtime,ctime,rtime,ttime,\n
					stats,FRONTEND,,,1,1,50000,13938,3693800,17784897,0,0,0,,,,,OPEN,,,,,,,,,1,2,0,,,,0,1,0,4,,,,0,13933,0,0,5,0,,1,4,13939,,,0,0,0,0,,,,,,,,\n
					stats,BACKEND,0,0,0,1,5000,5,3693800,17784897,0,0,,5,0,0,0,UP,0,0,0,,0,266274,0,,1,2,0,,0,,1,0,,3,,,,0,0,0,0,5,0,,,,,0,0,0,0,0,0,0,,,0,0,0,0,\n
					http_frontend,FRONTEND,,,0,0,50000,0,0,0,0,0,0,,,,,OPEN,,,,,,,,,1,3,0,,,,0,0,0,0,,,,0,0,0,0,0,0,,0,0,0,,,0,0,0,0,,,,,,,,\n
					unsecure,member:2:10.84.91.161,0,0,0,0,15000,0,0,0,,0,,0,0,0,0,UP,1,1,0,0,0,266274,0,,1,4,1,,0,,2,0,,0,L7OK,200,2,0,0,0,0,0,0,0,,,,0,0,,,,,-1,OK,,0,0,0,0,\n
					unsecure,BACKEND,0,0,0,0,5000,0,0,0,0,0,,0,0,0,0,UP,1,1,0,,0,266274,0,,1,4,0,,0,,1,0,,0,,,,0,0,0,0,0,0,,,,,0,0,0,0,0,0,-1,,,0,0,0,0,"),
			},
			TestDescription: "Successfully GET Haproxy status page",
		},
	}

	for _, test := range tests {
		g.Describe("getHaproxyStatus()", func() {
			g.It(test.TestDescription, func() {
				runner = test.HTTPRunner
				result := getHaproxyStatus(fakeConfig, make(chan []map[string]interface{}, 1))
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
