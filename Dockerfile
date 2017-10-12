FROM paas-docker-artifactory.gannettdigital.com/paas-newrelic-infra-base:latest
MAINTAINER PaaS-Delivery-API <paas-api@gannett.com>

RUN mkdir -p /var/db/newrelic-infra/custom-integrations/bin
RUN mkdir /var/db/newrelic-infra/integrations.d
COPY go-newrelic-plugin /var/db/newrelic-infra/custom-integrations/bin

# copy your config file to /var/db/newrelic-infa/custom-integrations
# copy your instance config file to /var/db/newrelic-infra/integrations.d
# run /usr/bin/newrelic-infra as CMD
