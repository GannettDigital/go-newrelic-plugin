package haproxy

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/GannettDigital/paas-api-utils/utilsHTTP"
	"github.com/Sirupsen/logrus"
)

var runner utilsHTTP.HTTPRunner

const NAME string = "haproxy"
const PROVIDER string = "haproxy"
const PROTOCOL_VERSION string = "1"

//HaproxyConfig is the keeper of the config
type HaproxyConfig struct {
	HaproxyPort      string
	HaproxyStatusURI string
	HaproxyHost      string
}

//BeStats holds all backend stats
type BeStats struct {
	QueueCurrent      int    `json:"qcur"`       //current queued requests.
	QueueMax          int    `json:"qmax"`       //max value of qcur
	SessionCurrent    int    `json:"scur"`       //current sessions
	SessionMax        int    `json:"smax"`       //max sessions
	SessionLimit      int    `json:"slim"`       //configured session limit
	SessionCumulative int    `json:"stot"`       //cumulative number of connections
	BytesInRate       int64  `json:"bin"`        //bytes in
	BytesOutRate      int64  `json:"bout"`       //bytes out
	DeniedReqRate     int    `json:"dreq"`       //requests denied because of security concerns.
	DeniedRespRate    int    `json:"dresp"`      //responses denied because of security concerns.
	ErrorsConRate     int    `json:"econ"`       //number of requests that encountered an error trying to connect to a backend server.
	ErrorsRespRate    int    `json:"eresp"`      //response errors. srv_abrt will be counted here also.
	WarnRedisRate     int    `json:"wredis"`     //number of times a request was redispatched to another server
	WarnRetrRate      int    `json:"wretr"`      //number of times a connection to a server was retried.
	Status            string `json:"status"`     //status (UP/DOWN/NOLB/MAINT/MAINT(via)...)
	Weight            int    `json:"weight"`     //total weight (backend), server weight (server)
	ActiveServers     int    `json:"act"`        //number of active servers (backend), server is active (server)
	BackupServers     int    `json:"bck"`        //number of backup servers (backend), server is backup (server)
	CheckDown         int    `json:"chkdown"`    //number of UP->DOWN transitions.
	LastChange        int    `json:"lastchg"`    //number of seconds since the last UP<->DOWN transition
	Downtime          int    `json:"downtime"`   //total downtime (in seconds).
	LBTotal           int    `json:"lbtot"`      //total number of times a server was selected, either for new sessions, or when re-dispatching
	SessionRate       int    `json:"rate"`       //number of sessions per second over last elapsed second
	SessionRateMax    int    `json:"rate_max"`   //max number of new sessions per second
	Resp1xx           int    `json:"hrsp_1xx"`   //http responses with 1xx code
	Resp2xx           int    `json:"hrsp_2xx"`   //http responses with 2xx code
	Resp3xx           int    `json:"hrsp_3xx"`   //http responses with 3xx code
	Resp4xx           int    `json:"hrsp_4xx"`   //http responses with 4xx code
	Resp5xx           int    `json:"hrsp_5xx"`   //http responses with 5xx code
	RespOther         int    `json:"hrsp_other"` //http responses with other codes (protocol error)
	ClientAborted     int    `json:"cli_abrt"`   //number of data transfers aborted by the client
	ServerAborted     int    `json:"srv_abrt"`   //number of data transfers aborted by the server
	QueueTime         int    `json:"qtime"`      //the average queue time in ms over the 1024 last requests
	ConnectTime       int    `json:"ctime"`      //the average connect time in ms over the 1024 last requests
	RespTime          int    `json:"rtime"`      //the average response time in ms over the 1024 last requests
	SessionTime       int    `json:"ttime"`      //the average total session time in ms over the 1024 last requests
}

//FeStats holds all frondend stats
type FeStats struct {
	SessionCurrent    int    `json:"scur"`       //current sessions
	SessionMax        int    `json:"smax"`       //max sessions
	SessionLimit      int    `json:"slim"`       //configured session limit
	SessionCumulative int    `json:"stot"`       //cumulative number of connections
	BytesInRate       int64  `json:"bin"`        //bytes in
	BytesOutRate      int64  `json:"bout"`       //bytes out
	DeniedReqRate     int    `json:"dreq"`       //requests denied because of security concerns.
	DeniedRespRate    int    `json:"dresp"`      //responses denied because of security concerns.
	ErrorsReqRate     int    `json:"ereq"`       //request errors.
	Status            string `json:"status"`     //status (UP/DOWN/NOLB/MAINT/MAINT(via)...)
	ReqRate           int    `json:"req_rate"`   //HTTP requests per second over last elapsed second
	Resp1xx           int    `json:"hrsp_1xx"`   //http responses with 1xx code
	Resp2xx           int    `json:"hrsp_2xx"`   //http responses with 2xx code
	Resp3xx           int    `json:"hrsp_3xx"`   //http responses with 3xx code
	Resp4xx           int    `json:"hrsp_4xx"`   //http responses with 4xx code
	Resp5xx           int    `json:"hrsp_5xx"`   //http responses with 5xx code
	RespOther         int    `json:"hrsp_other"` //http responses with other codes (protocol error)
	SessionRate       int    `json:"rate"`       //number of sessions per second over last elapsed second
}

//AllStats holds all haproxy stats
type AllStats struct {
	ProxyName               string `json:"pxname"`         //proxy name
	ServiceName             string `json:"svname"`         //svname: service name
	QueueCurrent            int    `json:"qcur"`           //current queued requests.
	QueueMax                int    `json:"qmax"`           //max value of qcur
	SessionCurrent          int    `json:"scur"`           //current sessions
	SessionMax              int    `json:"smax"`           //max sessions
	SessionLimit            int    `json:"slim"`           //configured session limit
	SessionCumulative       int    `json:"stot"`           //cumulative number of connections
	BytesInRate             int64  `json:"bin"`            //bytes in
	BytesOutRate            int64  `json:"bout"`           //bytes out
	DeniedReqRate           int    `json:"dreq"`           //requests denied because of security concerns.
	DeniedRespRate          int    `json:"dresp"`          //responses denied because of security concerns.
	ErrorsReqRate           int    `json:"ereq"`           //request errors.
	ErrorsConRate           int    `json:"econ"`           //number of requests that encountered an error trying to connect to a backend server.
	ErrorsRespRate          int    `json:"eresp"`          //response errors. srv_abrt will be counted here also.
	WarnRetrRate            int    `json:"wretr"`          //number of times a connection to a server was retried.
	WarnRedisRate           int    `json:"wredis"`         //number of times a request was redispatched to another server
	Status                  string `json:"status"`         //status (UP/DOWN/NOLB/MAINT/MAINT(via)...)
	Weight                  int    `json:"weight"`         //total weight (backend), server weight (server)
	ActiveServers           int    `json:"act"`            //number of active servers (backend), server is active (server)
	BackupServers           int    `json:"bck"`            //number of backup servers (backend), server is backup (server)
	CheckFailed             int    `json:"chkfail"`        //number of failed checks. (Only counts checks failed when the server is up.)
	CheckDown               int    `json:"chkdown"`        //number of UP->DOWN transitions.
	LastChange              int    `json:"lastchg"`        //number of seconds since the last UP<->DOWN transition
	Downtime                int    `json:"downtime"`       //total downtime (in seconds).
	QueueLimit              int    `json:"limit"`          //configured maxqueue for the server, or nothing in the value is 0
	PID                     int    `json:"pid"`            //process id (0 for first instance, 1 for second, ...)
	IID                     int    `json:"iid"`            //unique proxy id
	SID                     int    `json:"sid"`            //server id (unique inside a proxy)
	Throttle                int    `json:"throttle"`       //current throttle percentage for the server,
	LBTotal                 int    `json:"lbtot"`          //total number of times a server was selected, either for new sessions, or when re-dispatching
	TrackedID               int    `json:"tracked"`        //id of proxy/server if tracking is enabled.
	Type                    int    `json:"type"`           //(0=frontend, 1=backend, 2=server, 3=socket/listener)
	SessionRate             int    `json:"rate"`           //number of sessions per second over last elapsed second
	SessionRateLimit        int    `json:"rate_lim"`       //configured limit on new sessions per second
	SessionRateMax          int    `json:"rate_max"`       //max number of new sessions per second
	CheckStatus             string `json:"check_status"`   //status of last health check
	CheckCode               int    `json:"check_code"`     //layer5-7 code, if available
	CheckDuration           int    `json:"check_duration"` //time in ms took to finish last health check
	HTTPRsp1xx              int    `json:"hrsp_1xx"`       //http responses with 1xx code
	HTTPRsp2xx              int    `json:"hrsp_2xx"`       //http responses with 2xx code
	HTTPRsp3xx              int    `json:"hrsp_3xx"`       //http responses with 3xx code
	HTTPRsp4xx              int    `json:"hrsp_4xx"`       //http responses with 4xx code
	HTTPRsp5xx              int    `json:"hrsp_5xx"`       //http responses with 5xx code
	HTTPRsphrspother        int    `json:"hrsp_other"`     //http responses with other codes (protocol error)
	CheckFailDetails        int    `json:"hanafail"`       //failed health checks details
	HTTPRequestRate         int    `json:"req_rate"`       //HTTP requests per second over last elapsed second
	HTTPReqestRateMax       int    `json:"req_rate_max"`   //max number of HTTP requests per second observed
	HTTPReqestRateTot       int    `json:"req_tot"`        //total number of HTTP requests received
	ClientAborted           int    `json:"cli_abrt"`       //number of data transfers aborted by the client
	ServerAborted           int    `json:"srv_abrt"`       //number of data transfers aborted by the server
	CompressedBytesIn       int    `json:"comp_in"`        //number of HTTP response bytes fed to the compressor
	CompressedBytesOut      int    `json:"comp_out"`       //number of HTTP response bytes emitted by the compressor
	CompressedBytesBypassed int    `json:"comp_byp"`       //number of bytes that bypassed the HTTP compressor
	CompressedNumber        int    `json:"comp_rsp"`       //number of HTTP responses that were compressed
	SinceLastSession        int    `json:"lastsess"`       // number of seconds since last session assigned to sever/backend
	LastCheckDetails        string `json:"last_chk"`       //last health check contents or textual error
	LastAgentCheckDetails   int    `json:"last_agt"`       //last agent check contents or textual error
	QueueTime               int    `json:"qtime"`          //the average queue time in ms over the 1024 last requests
	ConnectTime             int    `json:"ctime"`          //the average connect time in ms over the 1024 last requests
	ResponseTime            int    `json:"rtime"`          //the average response time in ms over the 1024 last requests
	TotatlSessionTime       int    `json:"ttime"`          //the average total session time in ms over the 1024 last requests
}

// InventoryData is the data type for inventory data produced by a plugin data
// source and emitted to the agent's inventory data store
type InventoryData map[string]interface{}

// MetricData is the data type for events produced by a plugin data source and
// emitted to the agent's metrics data store
type MetricData map[string]interface{}

// EventData is the data type for single shot events
type EventData map[string]interface{}

// PluginData defines the format of the output JSON that plugins will return
type PluginData struct {
	Name            string                   `json:"name"`
	ProtocolVersion string                   `json:"protocol_version"`
	PluginVersion   string                   `json:"plugin_version"`
	Metrics         []MetricData             `json:"metrics"`
	Inventory       map[string]InventoryData `json:"inventory"`
	Events          []EventData              `json:"events"`
	Status          string                   `json:"status"`
}

// OutputJSON takes an object and prints it as a JSON string to the stdout.
// If the pretty attribute is set to true, the JSON will be idented for easy reading.
func OutputJSON(data interface{}, pretty bool) error {
	var output []byte
	var err error

	if pretty {
		output, err = json.MarshalIndent(data, "", "\t")
	} else {
		output, err = json.Marshal(data)
	}

	if err != nil {
		return fmt.Errorf("Error outputting JSON: %s", err)
	}

	if string(output) == "null" {
		fmt.Println("[]")
	} else {
		fmt.Println(string(output))
	}

	return nil
}

func Run(log *logrus.Logger, prettyPrint bool, version string) {

	// Initialize the output structure
	var data = PluginData{
		Name:            NAME,
		ProtocolVersion: PROTOCOL_VERSION,
		PluginVersion:   version,
		Inventory:       make(map[string]InventoryData),
		Metrics:         make([]MetricData, 0),
		Events:          make([]EventData, 0),
	}

	var config = HaproxyConfig{
		HaproxyPort:      os.Getenv("HAPROXYPORT"),
		HaproxyStatusURI: os.Getenv("HAPROXYSTATUSURI"),
		HaproxyHost:      os.Getenv("HAPROXYHOST"),
	}
	validateConfig(log, config)

	var metric = getHaproxyStatus(log, config)

	data.Metrics = append(data.Metrics, metric)
	fatalIfErr(log, OutputJSON(data, prettyPrint))
}

func initStats(log *logrus.Logger, config HaproxyConfig) (everything [][]string, err error) {
	haproxyStatsURI := fmt.Sprintf("%v:%v/%v;csv", config.HaproxyHost, config.HaproxyPort, config.HaproxyStatusURI)
	httpReq, err := http.NewRequest("GET", haproxyStatsURI, bytes.NewBuffer([]byte("")))
	if err != nil {
		log.WithFields(logrus.Fields{
			"haproxyStatsURI": haproxyStatsURI,
			"error":           err,
		}).Error("Encountered error creating http.NewRequest")
		close(stats)
		return []everything{}, err
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
		return err
	}
	r := csv.NewReader(strings.NewReader(string(data)))
	stats, err := r.ReadAll()
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("Unable to decode initial CSV stats")
		close(stats)
		return stats, err
	}
	return stats, nil
}

func getHaproxyStatus(log *logrus.Logger, config HaproxyConfig) (Stats []MetricData, err error) {
	InitialStats, err := initStats(log, config)
	if err != nil {
		log.WithFields(logrus.Fields{
			"haproxyConfig": config,
			"error":         err,
		}).Error("Encountered error querying Stats")
		return make([]MetricData, 0), err
	}
	Stats := make([]MetricData, 0)
	for _, record = range InitialStats {
		if record[0] == "http_frontend" {
			Stats = append(Stats, MetricData{
				"haproxy.frontend.session.current":  toInt(record[4]),
				"haproxy.frontend.session.max":      toInt(record[5]),
				"haproxy.frontend.session.limit":    toInt(record[6]),
				"haproxy.frontend.session.total":    toInt(record[7]),
				"haproxy.frontend.bytes.in_rate":    toInt64(record[8]),
				"haproxy.frontend.bytes.out_rate":   toInt64(record[9]),
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
			Stats = append(Stats, MetricData{
				"haproxy.backend.queue.current":       toInt(record[2]),
				"haproxy.backend.queue.max":           toInt(record[3]),
				"haproxy.backend.session.current":     toInt(record[4]),
				"haproxy.backend.session.max":         toInt(record[5]),
				"haproxy.backend.session.limit":       toInt(record[6]),
				"haproxy.backend.session.total":       toInt(record[7]),
				"haproxy.backend.bytes.in_rate":       toInt64(record[8]),
				"haproxy.backend.bytes.out_rate":      toInt64(record[9]),
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

func toInt64(value string) int64 {
	if value == "" {
		return 0
	} else {
		valueInt, err := strconv.ParseInt(value, 10, 64)
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

func validateConfig(log *logrus.Logger, config HaproxyConfig) {
	if config.HaproxyStatusURI == "" {
		log.Fatal("Config Yaml is missing values. Please check the config to continue")
	}
}

func fatalIfErr(log *logrus.Logger, err error) {
	if err != nil {
		log.WithError(err).Fatal("can't continue")
	}
}
