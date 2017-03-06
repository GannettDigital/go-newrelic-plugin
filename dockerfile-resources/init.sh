#!/bin/bash

# write $NR_KEY environment variable in place of '{NR_KEY}' in config.yaml if it exists
if [ -a /opt/gannett/newrelic/config.yaml ]; then
  sed -i "s/{NR_APP_NAME}/${NR_APP_NAME}/g" /opt/gannett/newrelic/config.yaml
  sed -i "s/{NR_KEY}/${NR_KEY}/g" /opt/gannett/newrelic/config.yaml
fi

/opt/gannett/newrelic/go-newrelic-plugin
