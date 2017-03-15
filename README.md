# go-newrelic-plugin

This repository holds the go-newrelic-plugin which uses combines collectors into one binary and makes them available to the newrelic-infra agent.

## How it Works

There are two parts to the architecture of the plugin.

### Commands
Commands live under the top level folder `cmd` These files are all apart of the same package `cmd` each command should have a corresponding collector(more on that later). We are using a package called  [cobra](https://github.com/spf13/cobra) to parse the commands and flags. This allows us t0 bundle all the collectors into one binary and also gives us an awesome help command.


```
-$ go run main.go help
A set of plugins to integrate custom checks into the newrelic infrastructure

Usage:
  go-newrelic-plugin [command]

Available Commands:
  couchbase   execute a couchbase collection
  help        Help about any command
  nginx       execute an nginx collection
  rabbitmq    execute a rabbitmq collection
  version     Print the version of go-newrelic-plugin

Flags:
      --pretty-print   pretty print output
      --verbose        verbose output
```

All of the commands besides [root.go]('./cmd/root.go') follow the same basic pattern. Import your collector and call the `Run` function. You can model your command function off of the skel.go command. Just make sure you update the `Use` and `Short` keys. `Use` is the name of the command and it should match the name of your collector. `Short` is a description of your collector. Both of these will show up in the help command output.

### Collectors

Collectors are designed to collect the stats for a given technology and report back to the newrelic infrastructure app. In general, collector development is where contributors will be spending their time.

Each collector is its own package. Take a look at the [skel package](''./skel/skel.go') The entry point to this package is `Run(log *logrus.Logger, prettyPrint bool, version string)`
Your collector's Run method will be called everytime New Relic requests stats.

Once your function is created, you can begin development of the logic for collecting and reporting stats of your specific technology.

##### Config
New Relic will pass in environment variables that you configure through a yaml config file. See ./skel/skel.yaml for an example
```
name: 'Name of your collector'
description: 'Short Description of your collector'
protocol_version: 'New Relics collection protocol'
os: 'Os this collector supports'

source:
  - command:
     - 'Location of your binary not really import for collector development'
    prefix: 'Application prefix for newrelic'
    interval: 'How often the agent checks for stats'
    env:
      KEY: "VALUE"
```
The important thing to note with the config is the env section. All of your config values should go here.

**Important**: In order to test your application you should export the variables you setup in your ~/.profile of ~/.bash_profile




### Standards
###### Naming
Your collector should be named after the technology you are gathering metrics for. If you were developing nginx, you collector would live in a file called `nginx.go` and live in a folder `nginx`

###### Errors
In the event that your collector has an error in retrieving stats and you are unable to report stats back, you should os.Exit(-1) or anything but zero to tell the newrelic agent their was an issue and to disregard any reported stats.


## Available Collectors
* [nginx]('./nginx/nginx.go')
* [rabbitmq]('./rabbitmq/rabbitmq.go')
* [couchbase]('./couchbase/couchbase.go')

#### Contributing

1. Create a go file, named after your collector. You can copy example.go as a starting point
2. Follow the function naming standards as defined above
3. Add config items to the config struct in `types.go`
4. Add your collector to the collector array in `types.go`
5. Update the following README sections
  - Available Collectors
  - Configuration example
6. Submit a PR
