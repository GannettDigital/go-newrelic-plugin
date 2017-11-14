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
	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/sirupsen/logrus"
)

type argumentList struct {
	sdkArgs.DefaultArgumentList
}

const (
	integrationName    = "com.gannettdigital.jira"
	integrationVersion = "0.1.0"
	jiraURL            = "https://jira.gannett.com"
)

var args argumentList

type Jira struct {
	Token      string
	Logger     *logrus.Logger
	MetricsSet map[string]JiraMetrics
}

type JiraMetrics struct {
	value      interface{}
	metricType int
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
	return &Jira{
		Token:  conf.authToken,
		Logger: logrus.New(),
	}, nil
}

func ValidateConfig(config Config) error {
	if config.authToken == "" {
		return errors.New("JIRA_AUTH_TOKEN must be set")
	}
	return nil
}

func (j *Jira) GetCurrentSprint() []int {
	resultSet := jiradata.SearchResults{}
	jreq := jiraRequest{
		method: "GET",
		uri:    "/rest/api/2/search",
		queryParams: map[string]string{
			"jql":    `sprint in openSprints() and PROJECT = "Platform as a Service (PAAS)"`,
			"fields": "customfield_10400",
		},
	}
	runner := &utilsHTTP.HTTPRunnerImpl{}
	_, b, err := j.executeJiraRequest(runner, jreq)
	if err != nil {
		j.Logger.Error(err)
	}

	if err = json.Unmarshal(b, &resultSet); err != nil {
		j.Logger.Error(err)
	}

	for _, i := range resultSet.Issues {
		if _, ok := i.Fields["customfield_10400"].([]interface{}); ok {
			d := i.Fields["customfield_10400"].([]interface{})[0].(string)
			r, _ := regexp.Compile(`id=(\d*)`)
			s := r.FindString(d)
			s = strings.Split(s, "=")[1]
			j.Logger.Infof("%+v", s)
		}
	}
	return []int{1}
}

func (j *Jira) GetOpenIssues(runner utilsHTTP.HTTPRunner) (jiradata.Issues, error) {
	resultSet := jiradata.SearchResults{}
	jreq := jiraRequest{
		method: "GET",
		uri:    "/rest/api/2/search",
		queryParams: map[string]string{
			"jql": fmt.Sprintf(`sprint in openSprints() and PROJECT = "PAAS"`),
			// customfield_10105 = storypoint,
			// customfield_11500 = epic
			"fields": "components, assignee, customfield_10105, customfield_11500",
		},
	}

	_, b, err := j.executeJiraRequest(runner, jreq)
	if err != nil {
		return resultSet.Issues, err
	}

	if err = json.Unmarshal(b, &resultSet); err != nil {
		return resultSet.Issues, err
	}
	// j.Logger.Printf("%v", resultSet)

	// for _, i := range resultSet.Issues {
	// 	if _, ok := i.Fields["components"].([]interface{}); ok {
	// 		components := i.Fields["components"].([]interface{})[0].(map[string]interface{})["name"].(string)
	// 		j.Logger.Info(components)
	// 	}
	// 	if _, ok := i.Fields["assignee"].(map[string]interface{}); ok {
	// 		assignee := i.Fields["assignee"].(map[string]interface{})["key"].(string)
	// 		j.Logger.Info("asginee:" + assignee)
	// 	}
	// 	if points, ok := i.Fields["customfield_10105"].(float64); ok {
	// 		j.Logger.Infof("points %d", int(points))
	// 	}
	// 	j.Logger.Infof("time spent %d", j.getWorkLogTotalTimeLogged(i.Key))
	// 	// j.Logger.Info
	// }
	// return len(resultSet.Issues), nil
	return resultSet.Issues, nil
}

func (j *Jira) getWorkLogTotalTimeLogged(id string) int {
	resultSet := jiradata.WorklogWithPagination{}

	jreq := jiraRequest{
		method: "GET",
		uri:    fmt.Sprintf("/rest/api/2/issue/%s/worklog", id),
	}

	runner := &utilsHTTP.HTTPRunnerImpl{}
	_, data, _ := j.executeJiraRequest(runner, jreq)

	err := json.Unmarshal(data, &resultSet)
	if err != nil {
		fmt.Println(err)
	}

	sumSeconds := 0
	for _, v := range resultSet.Worklogs {
		sumSeconds += v.TimeSpentSeconds
	}
	return sumSeconds

}

func (j *Jira) executeJiraRequest(runner utilsHTTP.HTTPRunner, jreq jiraRequest) (int, []byte, error) {
	jiraurl := fmt.Sprintf("%s%s", jiraURL, jreq.uri)
	req, err := http.NewRequest(jreq.method, jiraurl, nil)
	if err != nil {
		return 0, []byte{}, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", j.Token))

	params := url.Values{}
	for k, v := range jreq.queryParams {
		params.Add(k, v)
	}
	req.URL.RawQuery = params.Encode()

	return runner.CallAPI(j.Logger, nil, req, &http.Client{})
}

func Run(log *logrus.Logger, prettyPrint bool, version string) {
	j, err := NewJira(Config{authToken: os.Getenv("JIRA_TOKEN")})
	if err != nil {
		log.Print(err)
	}
	runner := &utilsHTTP.HTTPRunnerImpl{}
	// j.GetCurrentSprint()
	issues, err := j.GetOpenIssues(runner)
	if err != nil {
		log.Print(err)
	}
	log.Printf("%+v", issues)

	// // i, err := j.Progress(runner, 3760, "IN PROGRESS")
	// log.Print(i)

}
