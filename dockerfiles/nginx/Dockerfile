FROM paas-docker-artifactory.gannettdigital.com/paas-centos7-base:latest
MAINTAINER PaaS-Delivery <paas-delivery@gannett.com>

RUN mkdir -p /opt/gannett/newrelic

COPY go-newrelic-plugin config.yaml init.sh /opt/gannett/newrelic/

WORKDIR /opt/gannett/newrelic

CMD bash /opt/gannett/newrelic/init.sh
