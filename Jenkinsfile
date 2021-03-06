#!groovy

node {
  stage 'Execute Docker CI'
  withCredentials([
    string(credentialsId: "X_API_KEY", variable: "X_API_KEY"),
    string(credentialsId: "SCALR_KEY", variable: "SCALR_KEY"),
    string(credentialsId: "SLACK_URL", variable: "SLACK_URL"),
    string(credentialsId: "SCALR_SECRET_KEY", variable: "SCALR_SECRET_KEY"),
    string(credentialsId: "PASS_API_ART_KEY", variable: "PASS_API_ART_KEY"),
    string(credentialsId: "CONFIGS_ACCESS_KEY", variable: "CONFIGS_ACCESS_KEY"),
    string(credentialsId: "CONFIGS_SECRET_KEY", variable: "CONFIGS_SECRET_KEY"),
    string(credentialsId: "NEWRELIC_LICENSE_KEY", variable: "NEWRELIC_LICENSE_KEY"),
    string(credentialsId: "CODECOV_GO_NEWRELIC_PLUGIN", variable: "CODECOV_GO_NEWRELIC_PLUGIN"),
    string(credentialsId: "PAAS_API_CI_VAULT_TOKEN", variable: "PAAS_API_CI_VAULT_TOKEN")
  ]) {
    try {

      def paasApiCiVersion = "5.9.4-169"
      def repo = "GannettDigital/go-newrelic-plugin"
      def environment = "staging"
      def region = "us-east-1"
      def vaultURL = "https://vault.service.us-east-1.gciconsul.com:8200"
      def vaultConfig = "/secret/paas-api/paas-api-ci"

      print 'Running docker run'


      sh "docker run -e \"GIT_BRANCH=${env.BRANCH_NAME}\" -e \"VAULT_ADDR=${vaultURL}\" -e \"VAULT_CONFIG_LOCATION=${vaultConfig}\" -e \"VAULT_TOKEN=${PAAS_API_CI_VAULT_TOKEN}\" --rm -v ~/.docker/config.json:/root/.docker/config.json -v /var/run/docker.sock:/var/run/docker.sock paas-docker-artifactory.gannettdigital.com/paas-api-ci:${paasApiCiVersion} build \
        --repo=\"${repo}\" \
        --x-api-key=\"${X_API_KEY}\" \
        --x-scalr-access-key=\"${SCALR_KEY}\" \
        --x-scalr-secret-key=\"${SCALR_SECRET_KEY}\" \
        --x-kubernetes-api-user=\"not used\" \
        --x-kubernetes-api-token=\"not used\" \
        --slack-webhook=\"${SLACK_URL}\" \
        --artifactory-key=\"${PASS_API_ART_KEY}\" \
        --environment=\"${environment}\" \
        --region=\"${region}\" \
        --ci-job-number=${env.BUILD_ID} \
        --skip-deploy \
        --skip-swagger \
        --skip-source-check \
        --skip-validate \
        --newrelic-license=\"${NEWRELIC_LICENSE_KEY}\" \
        --codecov-token=\"${CODECOV_GO_NEWRELIC_PLUGIN}\""
    }
    catch (err) {
      currentBuild.result = "FAILURE"

      if (env.JOB_NAME.contains("master")) {
        def slackNotify = 'curl -X POST --data-urlencode \'payload= { "channel": "#api-releases", "username": "gopher-bot", "icon_emoji": ":httpmock:", "text": "*Build Failure*", "attachments": [ { "color": "#ff0000", "text": "Build failure for '
        slackNotify += env.JOB_NAME
        slackNotify += '","fields": [{"title": "Build URL","value": "'
        slackNotify += env.BUILD_URL
        slackNotify += '","short": true},{"title": "Build Number","value": "'
        slackNotify += env.BUILD_NUMBER
        slackNotify += '","short": true}]}]}\' '
        slackNotify += SLACK_URL

        sh slackNotify
      }

      throw err
    }
  }
}
