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
    string(credentialsId: "CODECOV_GO_NEWRELIC_PLUGIN", variable: "CODECOV_GO_NEWRELIC_PLUGIN"),
  ]) {
    try {

      def paasApiCiVersion = "3.27.1-96"
      def repo = "GannettDigital/go-newrelic-plugin"
      def environment = "staging"
      def region = "us-east-1"

      print 'Running docker run'


      sh "docker run -e \"GIT_BRANCH=${env.BRANCH_NAME}\" --rm -v /var/run/docker.sock:/var/run/docker.sock paas-docker-artifactory.gannettdigital.com/paas-api-ci:${paasApiCiVersion} build \
        --repo=\"${repo}\" \
        --x-api-key=\"${X_API_KEY}\" \
        --x-scalr-access-key=\"${SCALR_KEY}\" \
        --x-scalr-secret-key=\"${SCALR_SECRET_KEY}\" \
        --x-kubernetes-api-user=\"not used\" \
        --x-kubernetes-api-token=\"not used\" \
        --slack-webhook=\"${SLACK_URL}\" \
        --branch=\"${env.BRANCH_NAME}\" \
        --artifactory-key=\"${PASS_API_ART_KEY}\" \
        --aws-access-key-id=\"${CONFIGS_ACCESS_KEY}\" \
        --aws-secret-access-key=\"${CONFIGS_SECRET_KEY}\" \
        --environment=\"${environment}\" \
        --region=\"${region}\" \
        --ci-job-number=${env.BUILD_ID} \
        --skip-deploy \
        --skip-swagger \
        --skip-source-check \
        --skip-validate \
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
