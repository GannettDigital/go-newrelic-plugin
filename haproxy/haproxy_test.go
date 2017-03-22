package haproxy

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
		HaproxyPort:      "8000",
		HaproxyStatusURI: "haproxy",
		HaproxyHost:      "http://localhost",
	}
}

func TestinitStats(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		HTTPRunner      fake.HTTPResult
		TestDescription string
		ExpectedResult  [][]string
	}{
		{
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					{
						Method: "GET",
						URI:    "/haproxy;csv",
						Code:   200,
						Data: [][]string("# pxname,svname,qcur,qmax,scur,smax,slim,stot,bin,bout,dreq,dresp,ereq,econ,eresp,wretr,wredis,status,weight,act,bck,chkfail,chkdown,lastchg,downtime,qlimit,pid,iid,sid,throttle,lbtot,tracked,type,rate,rate_lim,rate_max,check_status,check_code,check_duration,hrsp_1xx,hrsp_2xx,hrsp_3xx,hrsp_4xx,hrsp_5xx,hrsp_other,hanafail,req_rate,req_rate_max,req_tot,cli_abrt,srv_abrt,comp_in,comp_out,comp_byp,comp_rsp,lastsess,last_chk,last_agt,qtime,ctime,rtime,ttime,",
							"stats,FRONTEND,,,1,2,50000,2572,682057,5657668,0,0,0,,,,,OPEN,,,,,,,,,1,2,0,,,,0,0,0,2,,,,0,2571,0,0,2,0,,1,3,2574,,,0,0,0,0,,,,,,,,",
							"stats,BACKEND,0,0,0,1,5000,2,682057,5657668,0,0,,2,0,0,0,UP,0,0,0,,0,49093,0,,1,2,0,,0,,1,0,,2,,,,0,0,0,0,2,0,,,,,0,0,0,0,0,0,0,,,861,0,0,1,",
							"http_frontend,FRONTEND,,,27,132,50000,68544,999168616,6150818383,0,0,429,,,,,OPEN,,,,,,,,,1,3,0,,,,0,3,0,24,,,,0,193685,729781,54761,2365,0,,20,90,980596,,,0,0,0,0,,,,,,,,",
							"unsecure,member:6:10.84.77.169,0,0,0,21,15000,196034,199530547,1227329342,,0,,0,0,0,0,UP,1,1,0,0,0,49093,0,,1,4,1,,196034,,2,4,,18,L7OK,200,5,0,38535,146150,10858,491,0,0,,,,1,0,,,,,0,OK,,0,2,137,816,",
							"unsecure,member:1:10.84.76.79,0,0,1,49,15000,196034,199553986,1227228046,,0,,0,0,0,0,UP,1,1,0,0,0,49093,0,,1,4,2,,196034,,2,4,,18,L7OK,200,4,0,38420,146278,10885,450,0,0,,,,2,0,,,,,0,OK,,0,2,161,796,",
							"unsecure,member:2:10.84.77.84,0,0,0,33,15000,196033,200057431,1232542076,,0,,0,1,0,0,UP,1,1,0,0,0,49093,0,,1,4,3,,196033,,2,3,,18,L7OK,200,43,0,38918,146002,10625,485,0,0,,,,2,0,,,,,0,OK,,0,2,158,643,",
							"unsecure,member:7:10.84.79.195,0,0,0,20,15000,196033,199664306,1233955111,,0,,0,0,0,0,UP,1,1,0,0,0,49093,0,,1,4,4,,196033,,2,4,,18,L7OK,200,4,0,38931,145677,10964,460,0,0,,,,3,0,,,,,0,OK,,0,2,141,694,",
							"unsecure,member:3:10.84.79.74,0,0,0,38,15000,196033,200351241,1229683585,,0,,0,1,0,0,UP,1,1,0,0,0,49093,0,,1,4,5,,196033,,2,4,,18,L7OK,200,7,0,38881,145674,11000,477,0,0,,,,2,0,,,,,0,OK,,0,2,196,844,",
							"unsecure,BACKEND,0,0,1,95,5000,980167,999157511,6150738160,0,0,,0,2,0,0,UP,5,5,0,,0,49093,0,,1,4,0,,980167,,1,20,,90,,,,0,193685,729781,54332,2365,0,,,,,10,0,0,0,0,0,0,,,0,2,169,837,"),
					},
				},
			},
			TestDescription: "Successfully GET HAProxy status page",
		},
	}

	for _, test := range tests {
		g.Describe("initStats()", func() {
			g.It(test.TestDescription, func() {
				runner = test.HTTPRunner
				result := initStats(logrus.New(), fakeConfig)
				g.Assert(reflect.DeepEqual(result, string(test.HTTPRunner.ResultsList[0].Data))).Equal(true)
			})
		})
	}
}

func TestgetHaproxyStatus(t *testing.T) {
	g := goblin.Goblin(t)

	result := map[string]interface{}{
		"haproxy.frontend.session.rate":       3,
		"haproxy.frontend.response.5xx":       2365,
		"haproxy.frontend.session.max":        132,
		"haproxy.frontend.bytes.in_rate":      999168616,
		"haproxy.frontend.response.other":     0,
		"haproxy.frontend.session.current":    27,
		"haproxy.frontend.response.3xx":       729781,
		"haproxy.frontend.errors.req_rate":    429,
		"haproxy.frontend.response.1xx":       0,
		"haproxy.frontend.requests.rate":      20,
		"haproxy.frontend.session.total":      68544,
		"haproxy.frontend.denied.req_rate":    0,
		"haproxy.frontend.denied.resp_rate":   0,
		"haproxy.frontend.response.2xx":       193685,
		"haproxy.frontend.response.4xx":       54761,
		"haproxy.frontend.session.limit":      50000,
		"haproxy.frontend.bytes.out_rate":     6150818383,
		"haproxy.backend.queue.current":       0,
		"haproxy.backend.session.limit":       5000,
		"haproxy.backend.errors.con_rate":     0,
		"haproxy.backend.response.3xx":        729781,
		"haproxy.backend.connect.time":        2,
		"haproxy.backend.response.2xx":        193685,
		"haproxy.backend.session.time":        837,
		"haproxy.backend.session.current":     1,
		"haproxy.backend.session.total":       980167,
		"haproxy.backend.bytes.in_rate":       999157511,
		"haproxy.backend.warnings.retr_rate":  0,
		"haproxy.backend.warnings.redis_rate": 0,
		"haproxy.backend.response.4xx":        54332,
		"haproxy.backend.response.5xx":        2365,
		"haproxy.backend.response.time":       169,
		"haproxy.backend.queue.max":           0,
		"haproxy.backend.session.max":         95,
		"haproxy.backend.denied.resp_rate":    0,
		"haproxy.backend.errors.resp_rate":    2,
		"haproxy.backend.session.rate":        20,
		"haproxy.backend.bytes.out_rate":      6150738160,
		"haproxy.backend.denied.req_rate":     0,
		"haproxy.backend.response.1xx":        0,
		"haproxy.backend.response.other":      0,
		"haproxy.backend.queue.time":          0,
	}

	var tests = []struct {
		Data            [][]string
		ExpectedResult  map[string]interface{}
		TestDescription string
	}{
		{
			Data:            "Active connections: 2 \nserver accepts handled requests\n 29 29 31 \nReading: 0 Writing: 1 Waiting: 1 ",
			ExpectedResult:  result,
			TestDescription: "Successfully parsed given stats",
		},
	}

	for _, test := range tests {
		g.Describe("TestgetHaproxyStatus()", func() {
			g.It(test.TestDescription, func() {
				result := TestgetHaproxyStatus(logrus.New(), test.Data)
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
			Value:           "8675301",
			ExpectedResult:  8675301,
			TestDescription: "Should return int 8675301 of string",
		},
		{
			Value:           "",
			ExpectedResult:  0,
			TestDescription: "Should return 0 if empty string",
		},
		{
			Value:           "arf",
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

func TestToInt64(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputValue      string
		ExpectedRes     int64
		TestDescription string
	}{
		{
			Value:           "98765432100",
			ExpectedResult:  98765432100,
			TestDescription: "Should return entire number when passed 98765432100 string",
		},
		{
			Value:           "",
			ExpectedResult:  0,
			TestDescription: "Should return 0 if empty string",
		},
		{
			Value:           "arf",
			ExpectedResult:  0,
			TestDescription: "Should return 0 if error converting to int",
		},
	}

	for _, test := range tests {
		g.Describe("toInt64()", func() {
			g.It(test.TestDescription, func() {
				result := toInt64(logrus.New(), test.Value)
				g.Assert(reflect.DeepEqual(result, test.ExpectedResult)).Equal(true)
			})
		})
	}
}
