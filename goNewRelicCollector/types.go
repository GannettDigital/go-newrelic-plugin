package goNewRelicCollector

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

// TODO:
// source from config.yaml

/*

Config example:

---
haproxy:
  enabled: true
  check_interval_ms: 100
nginx:
  enabled: true
  check_interval_ms: 1000
  listen_port: 80
  status_uri: /foo
  status_page: /derp
tags:
    region: us-east1
    farm: derp-farm

*/

type Config struct {
	AppName     string
	NginxConfig NginxConfig
}

type NginxConfig struct {
	NginxListenPort string
	NginxStatusURI  string
	NginxStatusPage string
}
