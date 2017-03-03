# go-newrelic-plugin

This repository holds the go-newrelic-plugin which uses [New Relic Go Agent](https://github.com/newrelic/go-agent) to send [custom events](https://docs.newrelic.com/docs/insights/new-relic-insights/custom-events/inserting-custom-events-new-relic-apm-agents) to Insights as a sort of plugin.

**How it Works**

The main loop instantiates a New Relic agent application with the specified [config](https://github.com/newrelic/go-agent/blob/master/config.go) object.  `WaitForConnection` is then called to allow the agent enough time to connect to New Relic before it starts collecting custom events.
Finally, a for loop is begun which does several things:
  * Passes the application to the `CustomEvent` function which contains the code for gathering your metrics data before it is sent as custom event via the `RecordCustomEvent` function.
  * Sleeps for a specified amount of time (default is one minute).  This controls how often your custom metrics are collected.  Please note the Agent will only send custom events to Insights every sixty seconds after `RecordCustomEvents` is called.
  * Keeps the plugin agent running until terminated by the system.

## Available Plugins
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

## Adding to this repository

Feel free to add custom metrics to this repo. Be sure to add a new event, test coverage, and update the README.
