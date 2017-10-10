FROM paas-docker-artifactory.gannettdigital.com/paas-centos7-base:latest

ARG VERSION=0.17.4-68

# go-new-relic
WORKDIR /opt/gannett
COPY go-newrelic-plugin /opt/gannet

# newrelic-infra install 
RUN echo "license_key: $NEWRELIC_LICENSE" | tee -a /etc/newrelic-infra.yml && \
    curl -o /etc/yum.repos.d/newrelic-infra.repo https://download.newrelic.com/infrastructure_agent/linux/yum/el/7/x86_64/newrelic-infra.repo && \
    yum -q makecache -y --disablerepo='*' --enablerepo='newrelic-infra' && \
    yum install newrelic-infra -y

ENTRYPOINT ["/opt/gannett/go-newrelic-plugin"]