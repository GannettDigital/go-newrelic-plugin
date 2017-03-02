package collectors

var CollectorArray map[string]Collector

// As we develop new collectors, add to this list here, so that our dispatcher
// knows about them
func init() {
	CollectorArray = map[string]Collector{
		"nginx":    NginxCollector,
		"rabbitmq": RabbitmqCollector,
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
type Collector func(config Config, stats chan<- []map[string]interface{})

type Config struct {
	AppName        string                  `yaml:"appname"`
	NewRelicKey    string                  `yaml:"newrelickey"`
	DefaultDelayMS int                     `yaml:"defaultdelayms"`
	Collectors     map[string]CommonConfig `yaml:"collectors"`
}

type CommonConfig struct {
	Enabled         bool            `yaml:"enabled" required:"true"`
	DelayMS         int             `yaml:"delayms"`
	CollectorConfig CollectorConfig `yaml:"collectorconfig"`
}

type CollectorConfig interface{}

type NginxConfig struct {
	NginxListenPort string `yaml:"nginxlistenport"`
	NginxStatusURI  string `yaml:"nginxstatusuri"`
	NginxStatusPage string `yaml:"nginxstatuspage"`
}

type RabbitmqConfig struct {
	RabbitMQUser     string `yaml:"rabbitmquser"`
	RabbitMQPassword string `yaml:"rabbitmqpassword"`
	RabbitMQPort     string `yaml:"rabbitmqport"`
	RabbitMQHost     string `yaml:"rabbitmqhost"`
}
