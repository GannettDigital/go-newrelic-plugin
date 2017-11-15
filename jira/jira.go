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

type argumentList struct {
	sdkArgs.DefaultArgumentList
}

const (
	integrationName    = "com.gannettdigital.jira"
	integrationVersion = "0.1.0"
	jiraURL            = "https://jira.gannett.com"
	metricSet          = "JiraMetrics"
)

var args argumentList

type Jira struct {
	Token  string
	Logger *logrus.Logger
}

type Config struct {
	authToken string
}

type jiraRequest struct {
	method      string
	uri         string
	queryParams map[string]string
}

func NewJira(conf Config) (*Jira, error) {
	if err := validateConfig(conf); err != nil {
		return nil, err
	}

	return &Jira{
		Token:  conf.authToken,
		Logger: logrus.New(),
	}, nil
}

func validateConfig(config Config) error {
	if config.authToken == "" {
		return errors.New("JIRA_AUTH_TOKEN must be set")
	}
	return nil
}

func processIssue(issue *jiradata.Issue, ms *metric.MetricSet) error {
	if _, ok := issue.Fields["components"].([]interface{}); ok {
		components := issue.Fields["components"].([]interface{})[0].(map[string]interface{})["name"].(string)
		ms.SetMetric("components", components, metric.ATTRIBUTE)
	}

	if assignee, ok := issue.Fields["assignee"].(map[string]interface{}); ok {
		assignee := assignee["key"].(string)
		ms.SetMetric("assignee", assignee, metric.ATTRIBUTE)
	}

	if points, ok := issue.Fields["customfield_10105"].(float64); ok {
		ms.SetMetric("storypoints", int(points), metric.GAUGE)
	}

	if _, ok := issue.Fields["customfield_10400"].([]interface{}); ok {
		d := issue.Fields["customfield_10400"].([]interface{})[0].(string)
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
			"jql": fmt.Sprintf(`sprint in openSprints() and PROJECT = "Platform as a Service (PAAS)"`),
			// customfield_10105 = storypoints,
			// customfield_11500 = epic
			// customfield_10400 = sprint number
			"maxResults": "200",
			"fields":     "components, assignee, customfield_10105, customfield_11500, customfield_10400",
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
	jiraurl := fmt.Sprintf("%s%s", jiraURL, jreq.uri)
	req, err := http.NewRequest(jreq.method, jiraurl, nil)
	if err != nil {
		return []byte{}, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", j.Token))

	params := url.Values{}
	for k, v := range jreq.queryParams {
		params.Add(k, v)
	}
	req.URL.RawQuery = params.Encode()

	code, b, err := runner.CallAPI(j.Logger, nil, req, &http.Client{})
	if code != 200 || err != nil {
		return []byte{}, errors.New("unable to grab jira data")
	}

	return b, nil
}

func Run(log *logrus.Logger) {
	conf := Config{
		authToken: os.Getenv("JIRA_TOKEN"),
	}

	runner := &utilsHTTP.HTTPRunnerImpl{}
	integration, err := sdk.NewIntegration(integrationName, integrationVersion, &args)
	if err != nil {
		log.Fatalf("unable to initialize new relic infrastracture, error: %s", err)
	}

	if err := emitMetrics(conf, runner, integration); err != nil {
		log.Fatalf("unable to emit metrics error: %s", err)
	}
}

type MetricEmmiter interface {
	NewMetricSet(string) *metric.MetricSet
	Publish() error
}

func emitMetrics(conf Config, runner utilsHTTP.HTTPRunner, integration MetricEmmiter) error {
	j, err := NewJira(conf)
	if err != nil {
		return err
	}

	searchResp, err := j.getOpenIssues(runner)
	if err != nil {
		return err
	}

	for _, i := range searchResp.Issues {
		ms := integration.NewMetricSet(metricSet)
		if err := processIssue(i, ms); err != nil {
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
