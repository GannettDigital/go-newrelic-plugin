package collectors

var CollectorArray map[string]Collector

// As we develop new collectors, add to this list here, so that our dispatcher
// knows about them
func init() {
	CollectorArray = map[string]Collector{
		"Nginx": NginxCollector,
	}

}

// a collector does some work to gather stats and
// returns a set of key => values that indicate metric type (key) and metric value
// its highly recommended that you namespace your metrics like <type>.<level1>.<leve2> etc
// ie: the haproxy Collector might return data as follows
/*
  map[string]string{
      "haproxy.beresp.500": "4",
      "haproxy.beresp.200": "100",
  }

*/
type Collector func(config Config, stats chan<- map[string]interface{})

type Config struct {
	AppName string
	Nginx   NginxConfig
}

type NginxConfig struct {
	Enabled         bool
	PollIntervalMS  int
	NginxListenPort string
	NginxStatusURI  string
	NginxStatusPage string
}
