# go-newrelic-plugin

This repository holds the go-newrelic-plugin which uses [New Relic Go Agent](https://github.com/newrelic/go-agent) to send [custom events](https://docs.newrelic.com/docs/insights/new-relic-insights/custom-events/inserting-custom-events-new-relic-apm-agents) to Insights. It was developed to address a feature gap between datadog and newrelic.

**How it Works**

There are two parts to the architecture of the plugin.

### Dispatcher

The dispatcher is the main component of the plugin. Its responsibilities include:
  - Reading from the config file
  - Setting up the connection to newrelirc
  - Orchestrating the timing and calls to the collectors (for enabled collectors)
  - Receiving results from a collector and reporting to newrelic
    - The Agent will only send custom events to Insights every sixty seconds after `RecordCustomEvents` is called.


The dispatcher is responsible for spinning off collector go routines on a timer and providing them a channel for them to report the stats back on. A new go routine is created per collector and per individual collection.

The dispatcher was designed to be developed once and should not need to be continually updated with the development of collectors.

### Collectors

Collectors are the work horses of the agent. They are designed to collect the stats for a given technology and report back to the dispatcher. In general, collector development is where contributors will be spending their time, and you generally don't need to know how the dispatcher works.

To become a collector, you must develop a method that is of type `Collector`. The following is the Collector Type.

```type Collector func(config Config, stats chan<- []map[string]interface{})```

In other words, you must create a function that matches a signature of `func(config Config, stats chan<- []map[string]interface{})`. Doing so allows your function to be of type `Collector` and can allow the dispatcher to hook into your code.

Once your function is created, you can begin development of the logic for collecting and reporting stats of your specific technology.

#### Arguments
##### Config

This is the main configuration, read by the dispatcher. You will have to get at your specific Collector Config like so:

```
var nginxConf NginxConfig
err := mapstructure.Decode(config.Collectors["nginx"].CollectorConfig, &nginxConf)
```

##### stats

The stats argument is a channel. Channels in go are used to exchange data between go routines. Here, it is used to communicate your gathered metrics back to the dispatcher, where they will be merged with tags and sent to newrelic. The channel is of type `[]map[string]interface{}`. Basically, this is an array of `hashes` (Key => Value pairs). If an individual event has multiple unique collections, those should each be separate items in your array. The dispatcher will send those up as unique newrelic events. An example of this is the rabbitmq collector, where each queue has its own stats.

Example of how to send metrics back to the collector using the stats channel:
```
stats <- []map[string]interface{}{
  {
    "nginx.net.connections": toInt(active),
    "nginx.net.accepts":     toInt(accepts),
    "nginx.net.handled":     toInt(handled),
    "nginx.net.requests":    toInt(requests),
    "nginx.net.writing":     toInt(writing),
    "nginx.net.waiting":     toInt(waiting),
    "nginx.net.reading":     toInt(reading),
  },
}```

** Important **

In the event that your collector has an error in retrieving stats and you are unable to report stats back, you must close the stats channel, to signal to the dispatcher that you had an error and the go routine should be cleaned up. Failing to do this, will leave go routines hanging around:

```
close(stats)
```

##### Standards
###### Naming
Your collector should be named after the technology you are gathering metrics for. If you were developing nginx, you collector would live in a file called `nginx.go` and your function would have a name of:

```
func NginxCollector(...
```
**Note:** It's important that you function name starts with a capital letter. In go, this signifies that the function is `exported` and can be used from external packages. The dispatcher will need this.


The metrics you report should be namespaced with the name of the collector:

```
"<collector>.stat1"
```

The dispatcher will report your metrics with an event name of:

```
gannettNewRelic<CollectorName>
```


## Available Collectors
* [nginx](#nginx)
* [rabbitmq](#rabbitmq)

## Configuration Examples

#### nginx

```yaml
nginx:
  enabled: false
  delayms: 1000
  collectorconfig:
    nginxlistenport: "8140"
    nginxstatusuri: nginx_status
    nginxstatuspage: http://localhost
```

#### rabbitMQ

```yaml
rabbitmq:
  enabled: true
  delayms: 1000
  collectorconfig:
    rabbitmquser: scalr
    rabbitmqpassword: secure
    rabbitmqport: "15672"
    rabbitmqhost: http://localhost
```


#### Contributing

1. Create a go file, named after your collector. You can copy example.go as a starting point
2. Follow the function naming standards as defined above
3. Add config items to the config struct in `types.go`
4. Add your collector to the collector array in `types.go`
5. Update the following README sections
  - Available Collectors
  - Configuration example
6. Submit a PR
