go-newrelic-plugin CHANGELOG
==============================

This file is used to list changes made in each version of go-newrelic-plugin.

# 0.13.0

Tom Barber - add additional scalrname for cluster name of couchbase cluster

# 0.12.1

Brian Lieberman - fix uninitialized logger in memcached plugin (PAAS-6521)


# 0.12.0

Michael Dunton - Mysql Plugin (PAAS-6237)

# 0.11.0

Michael Dunton - Memcached plugin added (PaaS-5495)

# 0.10.0

Vinod Vydier - added zookeeper plugin - uses zookeepers FLW commands conf and mntr (PAAS-5669)

# 0.9.3

Bridget Lane - Bump paas-api-ci

# 0.9.2

Seth Dozier- Haproxy multi-frontend hotfix

# 0.9.1

Bridget Lane - use credentials instead of globals

# 0.9.0

Tom Barber - adding support for replication set monitoring in mongo

# 0.8.2

Michael Dunton - More couchbase stats to use float32

# 0.8.1

Michael Dunton - Fixing bug in couchbase stats that was causing buckets to not show up in the metrics

# 0.8.0
Michael Dunton - Adding ssl Check Plugin.
              - Adding a ToInt to helpers

# 0.7.2

Ed De - PAAS-5763 Replace regexp FindString with FindStringSubmatch for Kraken
        Added Code Coverage reporting to Jenkinsfile

# 0.7.1

Tom Barber - update paas-api-ci version to start generating release notes

# 0.7.0

Ed De - PAAS-5120 Initial Kraken version

# 0.6.8

Tom Barber - update mongo plugin event_type to use correct value

# 0.6.7

Tom Barber - Remove no longer used dependencies and update paas-api-ci

# 0.6.6

Seth Dozier - PAAS-5255 cleanup empty metrics

# 0.6.5

Seth Dozier - PAAS-5255 cleanup config verbiage and fix pointer for successful haproxy integration tests

# 0.6.4

Alex Lindeman - PAAS-4898 change Jenkins event types to `CIJobSample` and `CIWorkerSample` and fix time units

# 0.6.3

Alex Lindeman - PAAS-4898 use helper `OutputJSON` method instead of own implementation in Jenkins collector

# 0.6.2

Michael Dunton - PAAS-5116 Updating readme with correct link to newrelic plugin spec

# 0.6.1

Michael Dunton - PAAS-5116 Updating readme and fixing event type issue on mongo

# 0.6.0

Seth Dozier - PAAS-4877 Initial HAProxy plugin

# 0.5.2

Michael Dunton - PAAS-5254 Add missing event_type for couchbase metrics

# 0.5.1

Michael Dunton - PAAS-5254 Fix event_type issue for couchbase

# 0.5.0

Michael Dunton  - PAAS-5131 Adding Mongo collector
                - Adding a helper file to do the outputJSON to reduce code copy.          

# 0.4.0

Alex Lindeman - PAAS-4898 Add Jenkins plugin

# 0.3.1

Ryan Grothouse - fixed bug with protocolVersion key

# 0.3.0

Tom Barber  - adding support for redis plugin

# 0.2.1

Michael Dunton  - PAAS-5087 Update Standards For contributing

# 0.2.0

Ryan Grothouse - refactor to use cobra for cli parsing of commands

# 0.1.1

Ryan Grothosue - bump paas-api-ci version

# 0.1.0

Ryan Grothouse - modified code to work with newrelics agent updates

# 0.0.20

Ryan Grothouse - remove standalone nginx code to its own repo

# 0.0.19

Michael Dunton - Adding prefix to yaml configs

# 0.0.18

Michael Dunton - Couchbase Standalone plugin for newrelic-infra

# 0.0.17

Michael Dunton - Rabbitmq Standalone plugin for newrelic-infra

# 0.0.16

Michael Dunton - Create a skeleton app for the standalone plugins

# 0.0.15

Michael Dunton - Creating standalone instance of nginx plugin for newrelic-infra sdk

# 0.0.14

Michael Dunton - Adding Couchbase support

# 0.0.13

Tom Barber - update default path to check for config

# 0.0.12

Ryan Grothouse - In-depth documentation on plugin archiecture, standards and contributing

# 0.0.11

Ryan Grothouse - randomize collector start time to make sure multiple collectors don't fire all at once

# 0.0.10

Tom Barber - add Dockerfile to build a docker container for go-newrelic-plugin

# 0.0.9

Tom Barber - add ability to attempt to load from s3 config when no local file is found

# 0.0.8

Tom Barber - allow optional global and collector specific key value and environment variable tags to be specified in config

# 0.0.7

Michael Dunton - Updating example.go and adding config examples to readme

# 0.0.6

Ryan Grothouse - catch panic so a poorly implemented collector can't nuke our monitor

# 0.0.5
Michael Dunton - adding support and Multiple Stats from a plugin and adding Rabbitmq

# 0.0.4
Bridget Lane - move utils.HTTPRunner for NGINX tests

# 0.0.3

Bridget Lane - NGINX report in ints, add testing

# 0.0.2

Tom Barber - adding support and structs for reading config file

# 0.0.1

Michael Dunton - initial change log commit of go-newrelic-plugin
