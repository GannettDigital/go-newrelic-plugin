package helpers

type NginxMetrics struct {
	Connections int //active connections
	Accepts     int //total accepts
	Handled     int //total handled connections
	Requests    int //total number of requests
	Writing     int //total number of connections where nginx is writing the response back to the client
	Waiting     int //current number of idle client connections waiting for a request
	Reading     int //connections where nginx is reading the request header
}

type NginxConfig struct {
	NginxListenPort string
	NginxStatusURI  string
	NginxStatusPage string
}
