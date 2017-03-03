#!/bin/bash

# write $NR_KEY environment variable in place of '{NR_KEY}' in site.conf
sed -i "s/{NR_APP_NAME}/${NR_APP_NAME}/g" /opt/gannett/newrelic/config.yaml
sed -i "s/{NR_KEY}/${NR_KEY}/g" /opt/gannett/newrelic/config.yaml
/opt/gannett/newrelic/go-newrelic-plugin
