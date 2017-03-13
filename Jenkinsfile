#!groovy

node {
  stage 'Execute Docker CI'
  try {

    def paasApiCiVersion = "1.10.0-14"
    def repo = "GannettDigital/go-newrelic-plugin"
    def environment = "staging"
    def region = "us-east-1"

    print 'Running docker run'

    sh "docker run -e \"GIT_BRANCH=${env.BRANCH_NAME}\" --rm -v /var/run/docker.sock:/var/run/docker.sock paas-docker-artifactory.gannettdigital.com/paas-api-ci:${paasApiCiVersion} build \
    --repo=\"${repo}\" \
    --x-api-key=\"${env.X_API_KEY}\" \
    --x-scalr-access-key=\"${env.SCALR_KEY}\" \
    --x-scalr-secret-key=\"${env.SCALR_SECRET_KEY}\" \
    --x-kubernetes-api-user=\"not used\" \
    --x-kubernetes-api-token=\"not used\" \
    --branch=\"${env.BRANCH_NAME}\" \
    --artifactory-key=\"${env.PASS_API_ART_KEY}\" \
    --aws-access-key-id=\"${env.CONFIGS_ACCESS_KEY}\" \
    --aws-secret-access-key=\"${env.CONFIGS_SECRET_KEY}\" \
    --environment=\"${environment}\" \
    --region=\"${region}\" \
    --ci-job-number=${env.BUILD_ID} \
    --skip-deploy \
    --skip-docker \ 
    --skip-swagger \
    --skip-source-check"
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
      slackNotify += env.SLACK_URL

      sh slackNotify
    }

    throw err
  }
}
