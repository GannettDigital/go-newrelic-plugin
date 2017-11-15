package jira

import (
	"errors"
	"os"
	"testing"

	"github.com/GannettDigital/paas-api-utils/utilsHTTP/fake"
	"github.com/franela/goblin"
	"github.com/newrelic/infra-integrations-sdk/metric"
)

func TestGetIssues(t *testing.T) {
	g := goblin.Goblin(t)
	j := NewJira(Config{authToken: "somefakeauth"})

	var tests = []struct {
		HTTPRunner      fake.HTTPResult
		TestDescription string
	}{
		{
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					{
						Method: "GET",
						URI:    "/rest/api/2/search?fields=components%2C+assignee%2C+customfield_10105%2C+customfield_11500%2C+customfield_10400&jql=sprint+in+openSprints%28%29+and+PROJECT+%3D+%22Platform+as+a+Service+%28PAAS%29%22&maxResults=200",
						Code:   200,
						Data:   issuesData(),
					},
				},
			},
			TestDescription: "Successfully GET jira issues",
		},
	}

	for _, test := range tests {
		g.Describe("GetJiraOpenIssues", func() {
			g.It(test.TestDescription, func() {
				runner := &test.HTTPRunner
				result, err := j.getOpenIssues(runner)
				g.Assert(err).Equal(nil)
				g.Assert(result.Total).Equal(123)
				g.Assert(len(result.Issues)).Equal(1)
				g.Assert(result.Issues[0].ID).Equal("354969")
			})
		})
	}
}

func TestGetWorkLogTotalTimeLogged(t *testing.T) {
	g := goblin.Goblin(t)
	j := NewJira(Config{authToken: "somefakeauth"})
	var tests = []struct {
		HTTPRunner      fake.HTTPResult
		TestDescription string
		ExpectedErr     error
		ExpectedOutput  int
	}{
		{
			TestDescription: "should succesfully get work log metrics",
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					{
						Method: "GET",
						URI:    "/rest/api/2/issue/PAAS-10283/worklog",
						Code:   200,
						Data:   tempoData(),
						Err:    nil,
					},
				},
			},
			ExpectedOutput: 25200,
		},
		{
			TestDescription: "should error when unable to reach jira",
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					{
						Method: "GET",
						URI:    "/rest/api/2/issue/PAAS-10283/worklog",
						Code:   500,
						Data:   []byte(``),
						Err:    errors.New("some jira error"),
					},
				},
			},
			ExpectedErr: errors.New("unable to grab jira data"),
		},
	}

	for _, test := range tests {
		result, err := j.getWorkLogTotalTimeLogged(&test.HTTPRunner, "PAAS-10283")
		g.Assert(result).Equal(test.ExpectedOutput)
		g.Assert(err).Equal(test.ExpectedErr)
	}
}

func TestValidateConfigs(t *testing.T) {
	g := goblin.Goblin(t)
	tests := []struct {
		SetConfig       func()
		ExpectedErr     error
		TestDescription string
	}{
		{
			SetConfig: func() {
				os.Setenv("JIRA_AUTH_TOKEN", "faketoken")
			},
			ExpectedErr:     errors.New("missing required config: [INTEGRATION_NAME INTEGRATION_VERSION JIRA_URL METRICSET_NAME]"),
			TestDescription: "should return error with missing fields",
		},
		{
			SetConfig: func() {
				os.Setenv("JIRA_AUTH_TOKEN", "faketoken")
				os.Setenv("INTEGRATION_NAME", "fakename")
				os.Setenv("INTEGRATION_VERSION", "fakeversion")
				os.Setenv("JIRA_URL", "fakeurl")
				os.Setenv("METRICSET_NAME", "fakemetricset")
			},
			TestDescription: "should not error all required fields are set",
		},
	}

	for _, test := range tests {
		g.Describe("validate configs", func() {
			g.It(test.TestDescription, func() {
				test.SetConfig()
				_, err := validateConfig()
				g.Assert(err).Equal(test.ExpectedErr)
			})
		})
	}

}

func TestEmitMetrics(t *testing.T) {
	g := goblin.Goblin(t)
	conf := Config{authToken: "something", metricSet: "JiraMetrics"}
	tests := []struct {
		HTTPRunner           fake.HTTPResult
		TestDescription      string
		InputEmitter         fakeEmitter
		ExpectedErr          error
		ExpectedMetricName   string
		ExpectedPublishCount int
		ExpectedNumberMetric int
	}{
		{
			TestDescription: "should successfully emit metrics",
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					{
						Method: "GET",
						URI:    "/rest/api/2/search?fields=components%2C+assignee%2C+customfield_10105%2C+customfield_11500%2C+customfield_10400&jql=sprint+in+openSprints%28%29+and+PROJECT+%3D+%22Platform+as+a+Service+%28PAAS%29%22&maxResults=200",
						Code:   200,
						Data:   issuesData(),
					},
					{
						Method: "GET",
						URI:    "/rest/api/2/issue/PAAS-10402/worklog",
						Code:   200,
						Data:   tempoData(),
					},
				},
			},
			InputEmitter:         fakeEmitter{},
			ExpectedMetricName:   "JiraMetrics",
			ExpectedPublishCount: 1,
			ExpectedNumberMetric: 6,
		},
		{
			TestDescription: "should error if unable to get jira issues",
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					{
						Method: "GET",
						URI:    "/rest/api/2/search?fields=components%2C+assignee%2C+customfield_10105%2C+customfield_11500%2C+customfield_10400&jql=sprint+in+openSprints%28%29+and+PROJECT+%3D+%22Platform+as+a+Service+%28PAAS%29%22&maxResults=200",
						Code:   500,
						Err:    errors.New("something"),
					},
				},
			},
			ExpectedErr:  errors.New("unable to grab jira data"),
			InputEmitter: fakeEmitter{},
		},
		{
			TestDescription: "should error unable to publish metrics",
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					{
						Method: "GET",
						URI:    "/rest/api/2/search?fields=components%2C+assignee%2C+customfield_10105%2C+customfield_11500%2C+customfield_10400&jql=sprint+in+openSprints%28%29+and+PROJECT+%3D+%22Platform+as+a+Service+%28PAAS%29%22&maxResults=200",
						Code:   200,
						Data:   issuesData(),
					},
					{
						Method: "GET",
						URI:    "/rest/api/2/issue/PAAS-10402/worklog",
						Code:   200,
						Data:   tempoData(),
					},
				},
			},
			ExpectedErr: errors.New("some publishing error"),
			InputEmitter: fakeEmitter{
				PublishErr: errors.New("some publishing error"),
			},
			ExpectedMetricName:   "JiraMetrics",
			ExpectedPublishCount: 1,
			ExpectedNumberMetric: 6,
		},
		{
			TestDescription: "should error if unable to get tempo issue",
			HTTPRunner: fake.HTTPResult{
				ResultsList: []fake.Result{
					{
						Method: "GET",
						URI:    "/rest/api/2/search?fields=components%2C+assignee%2C+customfield_10105%2C+customfield_11500%2C+customfield_10400&jql=sprint+in+openSprints%28%29+and+PROJECT+%3D+%22Platform+as+a+Service+%28PAAS%29%22&maxResults=200",
						Code:   200,
						Data:   issuesData(),
					},
					{
						Method: "GET",
						URI:    "/rest/api/2/issue/PAAS-10402/worklog",
						Code:   500,
						Err:    errors.New("somefakeerr"),
					},
				},
			},
			ExpectedErr:          errors.New("unable to grab jira data"),
			ExpectedMetricName:   "JiraMetrics",
			ExpectedNumberMetric: 5,
		},
	}
	for _, test := range tests {
		g.Describe("EmitMetrics", func() {
			g.It(test.TestDescription, func() {
				err := emitMetrics(conf, &test.HTTPRunner, &test.InputEmitter)
				g.Assert(err).Equal(test.ExpectedErr)
				g.Assert(test.InputEmitter.NewMetricSetNameCalled).Equal(test.ExpectedMetricName)
				g.Assert(test.InputEmitter.PublishCount).Equal(test.ExpectedPublishCount)
				if test.InputEmitter.MetricSet != nil {
					g.Assert(len(*test.InputEmitter.MetricSet)).Equal(test.ExpectedNumberMetric)
				}
			})
		})
	}
}

type fakeEmitter struct {
	PublishErr   error
	PublishCount int

	NewMetricSetNameCalled string
	MetricSet              *metric.MetricSet
}

func (f *fakeEmitter) Publish() error {
	f.PublishCount++
	return f.PublishErr
}

func (f *fakeEmitter) NewMetricSet(name string) *metric.MetricSet {
	f.NewMetricSetNameCalled = name
	metric := &metric.MetricSet{}
	f.MetricSet = metric

	return metric
}

func issuesData() []byte {
	return []byte(`
		{
			"expand":"names,schema",
			"startAt":0,
			"maxResults":1,
			"total":123,
			"issues":[
			   {
				  "expand":"operations,versionedRepresentations,editmeta,changelog,renderedFields",
				  "id":"354969",
				  "self":"fakeself",
				  "key":"PAAS-10402",
				  "fields":{
					 "customfield_10105":8.0,
					 "components":[
						{
						   "self":"fakecomponent",
						   "id":"26605",
						   "name":"success",
						   "description":"fakedescription"
						}
					 ],
					 "assignee":{
						"self":"fakeself",
						"name":"fake",
						"key":"paas-success",
						"emailAddress":"fake",
						"displayName":"display",
						"active":true,
						"timeZone":"America/New_York"
					 },
					 "customfield_11500":"display",
					 "customfield_10400":[  
						"com.atlassian.greenhopper.service.sprint.Sprint@3a896427[id=3772,rapidViewId=1779,state=ACTIVE,name=KoЯn,startDate=2017-11-08T14:00:42.550-06:00,endDate=2017-11-22T14:00:00.000-06:00,completeDate=<null>,sequence=3772]"
					 ]
				  }
			   }
			]
		 }
	`)
}

func tempoData() []byte {
	return []byte(`
		{  
			"startAt":0,
			"maxResults":1,
			"total":1,
			"worklogs":[  
			{  
				"self":"someworklog",
				"author":{  
					"self":"fake",
					"name":"fake",
					"key":"fake",
					"emailAddress":"fake",
					"avatarUrls":{},
					"displayName":"fake",
					"active":true,
					"timeZone":"America/New_York"
				},
				"updateAuthor":{ },
				"comment":"Working on issue PAAS-10283",
				"created":"2017-11-10T12:47:44.239-0600",
				"updated":"2017-11-10T12:47:44.239-0600",
				"started":"2017-11-10T00:00:00.000-0600",
				"timeSpent":"7h",
				"timeSpentSeconds":25200,
				"id":"212493",
				"issueId":"353718"
			}
			]
		}
	`)
}
