package collectors

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"
)

//BeStats holds all backend stats
type BeStats struct {
	BytesInRate    int //bytes in
	BytesOutRate   int //bytes out
	ConnectTime    int //ctime: the average connect time in ms over the 1024 last requests
	DeniedReqRate  int //dreq: requests denied because of security concerns.
	DeniedRespRate int //dresp: responses denied because of security concerns.
	ErrorsConRate  int //econ: number of requests that encountered an error trying to connect to a backend server.
	ErrorsRespRate int //eresp: response errors. srv_abrt will be counted here also.
	QueueCurrent   int //qcur: current queued requests.
	QueueTime      int //qtime: the average queue time in ms over the 1024 last requests
	Resp1xx        int //hrsp_1xx: http responses with 1xx code
	Resp2xx        int //hrsp_2xx: http responses with 2xx code
	Resp3xx        int //hrsp_3xx: http responses with 3xx code
	Resp4xx        int //hrsp_4xx: http responses with 4xx code
	Resp5xx        int //hrsp_5xx: http responses with 5xx code
	RespOther      int //hrsp_other: http responses with other codes (protocol error)
	RespTime       int // rtime: the average response time in ms over the 1024 last requests
	SessionCurrent int // scur: current sessions
	SessionLimit   int //slim: configured session limit
	SessionRate    int //rate: number of sessions per second over last elapsed second
	SessionTime    int //ttime: the average total session time in ms over the 1024 last requests
	WarnRedisRate  int //wredis: number of times a request was redispatched to another server
	WarnRetrRate   int //wretr: number of times a connection to a server was retried.
}

//FeStats holds all frondend stats
type FeStats struct {
	SessionCurrent    int    //scur: current sessions
	SessionMax        int    //smax: max sessions
	SessionLimit      int    //slim: configured session limit
	SessionCumulative int    // stot: cumulative number of connections
	BytesInRate       int    //bin: bytes in
	BytesOutRate      int    //bout: bytes out
	DeniedReqRate     int    //dreq: requests denied because of security concerns.
	DeniedRespRate    int    //dresp: responses denied because of security concerns.
	ErrorsReqRate     int    //ereq: request errors.
	Status            string // status: status (UP/DOWN/NOLB/MAINT/MAINT(via)...)
	ReqRate           int    //req_rate: HTTP requests per second over last elapsed second
	Resp1xx           int    //hrsp_1xx: http responses with 1xx code
	Resp2xx           int    //hrsp_2xx: http responses with 2xx code
	Resp3xx           int    //hrsp_3xx: http responses with 3xx code
	Resp4xx           int    //hrsp_4xx: http responses with 4xx code
	Resp5xx           int    //hrsp_5xx: http responses with 5xx code
	RespOther         int    //hrsp_other: http responses with other codes (protocol error)
	SessionRate       int    //rate: number of sessions per second over last elapsed second
}

//AllStats holds all haproxy stats
type AllStats struct {
	pxname        string
	svname        string //svname: service name
	qcur          int    //qcur: current queued requests.
	qmax          int    //qmax: max value of qcur
	scur          int    //scur: current sessions
	smax          int    //smax: max sessions
	slim          int    //slim: configured session limit
	stot          int    // stot: cumulative number of connections
	bin           int    //bin: bytes in
	bout          int    //bout: bytes out
	dreq          int    //dreq: requests denied because of security concerns.
	dresp         int    //dresp: responses denied because of security concerns.
	ereq          int    //ereq: request errors.
	econ          int    //econ: number of requests that encountered an error trying to connect to a backend server.
	eresp         int    //eresp: response errors. srv_abrt will be counted here also.
	wretr         int    //wretr: number of times a connection to a server was retried.
	wredis        int    //wredis: number of times a request was redispatched to another server
	status        string // status: status (UP/DOWN/NOLB/MAINT/MAINT(via)...)
	weight        int    // weight: total weight (backend), server weight (server)
	act           int    //act: number of active servers (backend), server is active (server)
	bck           int    //bck: number of backup servers (backend), server is backup (server)
	chkfail       int    //chkfail: number of failed checks. (Only counts checks failed when the server is up.)
	chkdown       int    //chkdown: number of UP->DOWN transitions.
	lastchg       int    // lastchg: number of seconds since the last UP<->DOWN transition
	downtime      int    //downtime: total downtime (in seconds).
	qlimit        int    //limit: configured maxqueue for the server, or nothing in the value is 0
	pid           int    //pid: process id (0 for first instance, 1 for second, ...)
	iid           int    // iid: unique proxy id
	sid           int    //sid ; server id (unique inside a proxy)
	throttle      int    //throttle ; current throttle percentage for the server,
	lbtot         int    //lbtot: total number of times a server was selected, either for new sessions, or when re-dispatching
	tracked       int    //tracked ; id of proxy/server if tracking is enabled.
	xtype         int    //type: (0=frontend, 1=backend, 2=server, 3=socket/listener)
	rate          int    //rate ; number of sessions per second over last elapsed second
	ratelim       int    //rate_lim ; configured limit on new sessions per second
	ratemax       int    //rate_max ; max number of new sessions per second
	checkstatus   string //check_status ; status of last health check
	checkcode     int    //check_code ; layer5-7 code, if available
	checkduration int    //check_duration ; time in ms took to finish last health check
	hrsp1xx       int    // hrsp_1xx ; http responses with 1xx code
	hrsp2xx       int    //hrsp_2xx ; http responses with 2xx code
	hrsp3xx       int    // hrsp_3xx ; http responses with 3xx code
	hrsp4xx       int    //hrsp_4xx ; http responses with 4xx code
	hrsp5xx       int    //hrsp_5xx ; http responses with 5xx code
	hrspother     int    // hrsp_other ; http responses with other codes (protocol error)
	hanafail      int    //hanafail ; failed health checks details
	reqrate       int    //req_rate ; HTTP requests per second over last elapsed second
	reqratemax    int    // req_rate_max ; max number of HTTP requests per second observed
	reqtot        int    // req_tot ; total number of HTTP requests received
	cliabrt       int    // cli_abrt: number of data transfers aborted by the client
	srvabrt       int    // srv_abrt: number of data transfers aborted by the server
	compin        int    // comp_in ; number of HTTP response bytes fed to the compressor
	compout       int    // comp_out ; number of HTTP response bytes emitted by the compressor
	compbyp       int    // comp_byp ; number of bytes that bypassed the HTTP compressor
	comprsp       int    //comp_rsp ; number of HTTP responses that were compressed
	lastsess      int    // number of seconds since last session assigned to sever/backend
	lastchk       string // last_chk ; last health check contents or textual error
	lastagt       int    // last_agt ; last agent check contents or textual error
	qtime         int    // qtime: the average queue time in ms over the 1024 last requests
	ctime         int    // ctime: the average connect time in ms over the 1024 last requests
	rtime         int    // rtime: the average response time in ms over the 1024 last requests
	ttime         int    // ttime: the average total session time in ms over the 1024 last requests
}

func iStats(config HaproxyConfig) (everything []AllStats, err error) {
	haproxyStatsURI := fmt.Sprintf("%v:%v/%v;csv", config.HaproxyHost, config.HaproxyPort, config.HaproxyStatusURI)
	httpReq, err := http.NewRequest("GET", haproxyStatsURI, bytes.NewBuffer([]byte("")))
	if err != nil {
		log.WithFields(logrus.Fields{
			"haproxyStatsURI": haproxyStatsURI,
			"error":           err,
		}).Error("Encountered error creating http.NewRequest")
		close(stats)
		return ""
	}
	code, data, err := runner.CallAPI(log, nil, httpReq, &http.Client{})
	if err != nil || code != 200 {
		log.WithFields(logrus.Fields{
			"code":    code,
			"data":    string(data),
			"httpReq": httpReq,
			"error":   err,
		}).Error("Encountered error calling CallAPI")
		close(stats)
		return ""
	}
	r := csv.NewReader(strings.NewReader(string(data)))
	stats, err := r.ReadAll()
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("Unable to decode initial CSV stats")
		close(stats)
		return ""
	}
	return stats
}

func getHaproxyStatus(config HaproxyConfig) ([]map[string]interface{}, error) {
	InitialStats, err := iStats(config)
	if err != nil {
		log.WithFields(logrus.Fields{
			"haproxyConfig": config,
			"error":         err,
		}).Error("Encountered error querying Stats")
		return make([]map[string]interface{}, 0), err
	}
	Stats := make([]map[string]interface{}, 0)
	for _, record = range InitialStats {
		if record[0] == "http_frontend" {
			Stats = append(Stats, map[string]interface{}{
				"haproxy.frontend.session.current":  toInt(record[4]),
				"haproxy.frontend.session.max":      toInt(record[5]),
				"haproxy.frontend.session.limit":    toInt(record[6]),
				"haproxy.frontend.session.total":    toInt(record[7]),
				"haproxy.frontend.bytes.in_rate":    toInt(record[8]),
				"haproxy.frontend.bytes.out_rate":   toInt(record[9]),
				"haproxy.frontend.denied.req_rate":  toInt(record[10]),
				"haproxy.frontend.denied.resp_rate": toInt(record[11]),
				"haproxy.frontend.errors.req_rate":  toInt(record[12]),
				"haproxy.frontend.session.rate":     toInt(record[33]),
				"haproxy.frontend.response.1xx":     toInt(record[39]),
				"haproxy.frontend.response.2xx":     toInt(record[40]),
				"haproxy.frontend.response.3xx":     toInt(record[41]),
				"haproxy.frontend.response.4xx":     toInt(record[42]),
				"haproxy.frontend.response.5xx":     toInt(record[43]),
				"haproxy.frontend.response.other":   toInt(record[44]),
				"haproxy.frontend.requests.rate":    toInt(record[46]),
			})
		} else if record[0] != "stats" && record[1] == "BACKEND" {
			Stats = append(Stats, map[string]interface{}{
				"haproxy.backend.queue.current":       toInt(record[2]),
				"haproxy.backend.queue.max":           toInt(record[3]),
				"haproxy.backend.session.current":     toInt(record[4]),
				"haproxy.backend.session.max":         toInt(record[5]),
				"haproxy.backend.session.limit":       toInt(record[6]),
				"haproxy.backend.session.total":       toInt(record[7]),
				"haproxy.backend.bytes.in_rate":       toInt(record[8]),
				"haproxy.backend.bytes.out_rate":      toInt(record[9]),
				"haproxy.backend.denied.req_rate":     toInt(record[10]),
				"haproxy.backend.denied.resp_rate":    toInt(record[11]),
				"haproxy.backend.errors.con_rate":     toInt(record[13]),
				"haproxy.backend.errors.resp_rate":    toInt(record[14]),
				"haproxy.backend.warnings.retr_rate":  toInt(record[15]),
				"haproxy.backend.warnings.redis_rate": toInt(record[16]),
				"haproxy.backend.session.rate":        toInt(record[33]),
				"haproxy.backend.response.1xx":        toInt(record[39]),
				"haproxy.backend.response.2xx":        toInt(record[40]),
				"haproxy.backend.response.3xx":        toInt(record[41]),
				"haproxy.backend.response.4xx":        toInt(record[42]),
				"haproxy.backend.response.5xx":        toInt(record[43]),
				"haproxy.backend.response.other":      toInt(record[44]),
				"haproxy.backend.queue.time":          toInt(record[58]),
				"haproxy.backend.connect.time":        toInt(record[59]),
				"haproxy.backend.response.time":       toInt(record[60]),
				"haproxy.backend.session.time":        toInt(record[61]),
			})
		}
	}
	//return Stats, nil
	return Stats, nil
}

func toInt(value string) int {
	if value == "" {
		return 0
	} else {
		valueInt, err := strconv.Atoi(value)
		if err != nil {
			log.WithFields(logrus.Fields{
				"valueInt": valueInt,
				"error":    err,
			}).Error("Error converting value to int")

			return 0
		}

		return valueInt
	}
}

// HaproxyCollector used for reference for collector developers
func HaproxyCollector(config Config, stats chan<- []map[string]interface{}) {
	var haproxyconf HaproxyConfig
	err := mapstructure.Decode(config.Collectors["haproxy"].CollectorConfig, &haproxyconf)
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("Unable to decode HAProxy config into HaproxyConfig object")
		close(stats)
	}
	lbStatus, getStatsError := getHaproxyStatus(haproxyconf)
	if getStatsError != nil {
		log.WithFields(logrus.Fields{
			"err": getStatsError,
		}).Error("Error retreiving haproxy stats.")
		close(stats)
		return
	}
	stats <- lbStatus
}
