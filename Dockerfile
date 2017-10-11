FROM paas-docker-artifactory.gannettdigital.com/paas-newrelic-infra-base:latest
MAINTAINER PaaS-Delivery-API <paas-api@gannett.com>

RUN mkdir /opt/gannett
COPY go-newrelic-plugin /opt/gannett

ENTRYPOINT ["/opt/gannett/go-newrelic-plugin"]
