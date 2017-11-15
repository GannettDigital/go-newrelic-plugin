package jira

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/GannettDigital/paas-api-utils/utilsHTTP"

	"github.com/Netflix-Skunkworks/go-jira/jiradata"
	"github.com/Sirupsen/logrus"
	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/newrelic/infra-integrations-sdk/sdk"
)

const (
	projectName      = "Platform as a Service (PAAS)"
	storyPointsField = "customfield_10105"
	epicField        = "customfield_11500"
	sprintIDField    = "customfield_10400"
)

type argumentList struct {
	sdkArgs.DefaultArgumentList
}

var args argumentList

type Jira struct {
	Token      string
	URL        string
	Logger     *logrus.Logger
	HTTPClient *http.Client
}

// Config holds Jira and NewRelic Configs
type Config struct {
	authToken          string
	integrationName    string
	integrationVersion string
	jiraURL            string
	metricSet          string
}

type jiraRequest struct {
	method      string
	uri         string
	queryParams map[string]string
}

// NewJira returns a new Jira
func NewJira(conf Config) *Jira {
	return &Jira{
		Token:      conf.authToken,
		URL:        conf.jiraURL,
		HTTPClient: &http.Client{},
		Logger:     logrus.New(),
	}
}

func validateConfig(config Config) error {
	missingFields := make([]string, 0)

	if config.authToken == "" {
		missingFields = append(missingFields, "JIRA_AUTH_TOKEN")
	}
	if config.integrationName == "" {
		missingFields = append(missingFields, "NR_INTEGRATION_NAME")
	}
	if config.integrationVersion == "" {
		missingFields = append(missingFields, "NR_INTEGRATION_VERSION")
	}
	if config.jiraURL == "" {
		missingFields = append(missingFields, "JIRA_URL")
	}
	if config.metricSet == "" {
		missingFields = append(missingFields, "NR_METRICSET_NAME")
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("missing required config: %v", missingFields)
	}

	return nil
}

// extractIssueStats grabs data from a Jira issue and adds to a MetricSet
func extractIssueStats(issue *jiradata.Issue, ms *metric.MetricSet) error {
	if _, ok := issue.Fields["components"].([]interface{}); ok {
		components := issue.Fields["components"].([]interface{})[0].(map[string]interface{})["name"].(string)
		ms.SetMetric("components", components, metric.ATTRIBUTE)
	}

	if assignee, ok := issue.Fields["assignee"].(map[string]interface{}); ok {
		assignee := assignee["key"].(string)
		ms.SetMetric("assignee", assignee, metric.ATTRIBUTE)
	}

	if points, ok := issue.Fields[storyPointsField].(float64); ok {
		ms.SetMetric("storypoints", int(points), metric.GAUGE)
	}

	if _, ok := issue.Fields[sprintIDField].([]interface{}); ok {
		d := issue.Fields[sprintIDField].([]interface{})[0].(string)
		r, err := regexp.Compile(`id=(\d*)`)
		if err != nil {
			return err
		}
		var id string
		s := strings.Split(r.FindString(d), "=")
		if len(s) >= 1 {
			id = s[1]
		}
		ms.SetMetric("sprintID", id, metric.ATTRIBUTE)
	}

	ms.SetMetric("storyID", issue.Key, metric.ATTRIBUTE)
	return nil
}

func (j *Jira) getOpenIssues(runner utilsHTTP.HTTPRunner) (jiradata.SearchResults, error) {
	resultSet := jiradata.SearchResults{}
	jreq := jiraRequest{
		method: "GET",
		uri:    "/rest/api/2/search",
		queryParams: map[string]string{
			"jql":        fmt.Sprintf(`sprint in openSprints() and PROJECT = %q`, projectName),
			"maxResults": "200",
			"fields":     fmt.Sprintf("components, assignee, %s, %s, %s", storyPointsField, epicField, sprintIDField),
		},
	}

	b, err := j.executeJiraRequest(runner, jreq)
	if err != nil {
		return resultSet, err
	}

	if err = json.Unmarshal(b, &resultSet); err != nil {
		return resultSet, err
	}

	return resultSet, nil
}

func (j *Jira) getWorkLogTotalTimeLogged(runner utilsHTTP.HTTPRunner, storyID string) (int, error) {
	resultSet := jiradata.WorklogWithPagination{}

	jreq := jiraRequest{
		method: "GET",
		uri:    fmt.Sprintf("/rest/api/2/issue/%s/worklog", storyID),
	}

	b, err := j.executeJiraRequest(runner, jreq)
	if err != nil {
		return 0, err
	}

	if err = json.Unmarshal(b, &resultSet); err != nil {
		return 0, err
	}

	seconds := 0
	for _, v := range resultSet.Worklogs {
		seconds += v.TimeSpentSeconds
	}
	return seconds, nil
}

func (j *Jira) executeJiraRequest(runner utilsHTTP.HTTPRunner, jreq jiraRequest) ([]byte, error) {
	jiraURL := fmt.Sprintf("%s%s", j.URL, jreq.uri)
	req, err := http.NewRequest(jreq.method, jiraURL, nil)
	if err != nil {
		return []byte{}, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", j.Token))

	params := url.Values{}
	for k, v := range jreq.queryParams {
		params.Add(k, v)
	}
	req.URL.RawQuery = params.Encode()

	code, b, err := runner.CallAPI(j.Logger, nil, req, j.HTTPClient)
	if code != 200 || err != nil {
		return []byte{}, errors.New("unable to grab jira data")
	}

	return b, nil
}

func Run(log *logrus.Logger) {
	conf := Config{
		authToken:          os.Getenv("JIRA_AUTH_TOKEN"),
		integrationName:    os.Getenv("NR_INTEGRATION_NAME"),
		integrationVersion: os.Getenv("NR_INTEGRATION_VERSION"),
		jiraURL:            os.Getenv("JIRA_URL"),
		metricSet:          os.Getenv("NR_METRICSET_NAME"),
	}
	if err := validateConfig(conf); err != nil {
		log.Fatal(err)
	}

	runner := &utilsHTTP.HTTPRunnerImpl{}
	integration, err := sdk.NewIntegration(conf.integrationName, conf.integrationVersion, &args)
	if err != nil {
		log.Fatalf("unable to initialize new relic infrastracture, error: %s", err)
	}

	if err := emitMetrics(conf, runner, integration); err != nil {
		log.Fatalf("unable to emit metrics error: %s", err)
	}
}

// MetricEmiter registers metrics and flushes metrics to standard out
type MetricEmiter interface {
	NewMetricSet(string) *metric.MetricSet
	Publish() error
}

func emitMetrics(conf Config, runner utilsHTTP.HTTPRunner, integration MetricEmiter) error {
	j := NewJira(conf)

	searchResp, err := j.getOpenIssues(runner)
	if err != nil {
		return err
	}

	for _, i := range searchResp.Issues {
		ms := integration.NewMetricSet(conf.metricSet)
		if err := extractIssueStats(i, ms); err != nil {
			return err
		}

		tempo, err := j.getWorkLogTotalTimeLogged(runner, i.Key)
		if err != nil {
			return err
		}
		ms.SetMetric("tempo.seconds", tempo, metric.GAUGE)
		if err := integration.Publish(); err != nil {
			return err
		}
	}
	return nil
}
