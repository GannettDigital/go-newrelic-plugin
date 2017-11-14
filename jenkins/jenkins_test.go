package jenkins

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/bndr/gojenkins"
	"github.com/franela/goblin"
	"github.com/jarcoal/httpmock"
	"github.com/sirupsen/logrus"
)

var (
	fakeLog    = logrus.New()
	fakeConfig = Config{
		JenkinsHost:    "http://jenkins.mock",
		JenkinsAPIUser: "test-user",
		JenkinsAPIKey:  "test-pw",
	}
)

func TestValidateConfig(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("jenkins validateConfig()", func() {
		expected := map[string]struct {
			ExpectedIsNil bool
			Config        Config
		}{
			"no":                            {false, Config{}},
			"JenkinsHost":                   {true, Config{JenkinsHost: "http://jenkins.mock"}},
			"JenkinsHost, JenkinsAPIUser":   {false, Config{JenkinsHost: "http://jenkins.mock", JenkinsAPIUser: "test-user"}},
			"JenkinsHost, JenkinsAPIKey":    {false, Config{JenkinsHost: "http://jenkins.mock", JenkinsAPIKey: "test-pw"}},
			"JenkinsAPIUser, JenkinsAPIKey": {false, Config{JenkinsAPIUser: "test-user", JenkinsAPIKey: "test-pw"}},
			"all": {true, Config{JenkinsHost: "http://jenkins.mock", JenkinsAPIUser: "test-user", JenkinsAPIKey: "test-pw"}},
		}
		for name, ex := range expected {
			desc := fmt.Sprintf("should return %v when %v fields are set", ex.ExpectedIsNil, name)
			g.It(desc, func() {
				valid := validateConfig(ex.Config)
				g.Assert(valid == nil).Equal(ex.ExpectedIsNil)
			})
		}
	})
}

func TestGetJenkins(t *testing.T) {
	g := goblin.Goblin(t)
	fakeJenkins := fakeJenkins()
	g.Describe("jenkins getJenkins()", func() {
		res := getJenkins(fakeConfig)
		g.It("should connect to the right Jenkins", func() {
			g.Assert(res.Server).Equal("http://jenkins.mock")
		})
		g.It("should have authorization data", func() {
			g.Assert(res.Requester.BasicAuth.Username).Equal("test-user")
			g.Assert(res.Requester.BasicAuth.Password).Equal("test-pw")
		})
		g.It("should be using httpmock.MockTransport for requests while testing", func() {
			transport := reflect.TypeOf(fakeJenkins.Requester.Client.Transport)
			g.Assert(transport.String()).Equal("*httpmock.MockTransport")
		})
	})
}

func TestGetMetrics(t *testing.T) {
	g := goblin.Goblin(t)
	fakeJenkins := fakeJenkins()
	g.Describe("jenkins getMetrics()", func() {
		res, err := getMetrics(fakeLog, fakeJenkins)
		g.It("should return metric data", func() {
			g.Assert(err).Equal(nil)
			g.Assert(len(res) > 0).Equal(true)
		})
		g.It("should have 'event_type' keys on everything", func() {
			for _, metric := range res {
				g.Assert(metric["event_type"] != nil).Equal(true)
			}
		})
		g.It("should have 'entity_name' keys on everything", func() {
			for _, metric := range res {
				g.Assert(metric["entity_name"] != nil).Equal(true)
			}
		})
		g.It("should have 'provider' keys on everything", func() {
			for _, metric := range res {
				g.Assert(metric["provider"] != nil).Equal(true)
			}
		})
	})
}

func TestGetJobStats(t *testing.T) {
	g := goblin.Goblin(t)
	fakeJenkins := fakeJenkins()
	g.Describe("jenkins getJobStats()", func() {
		expected := map[string]JobMetric{
			"with build and test data": {
				EntityName:      "foo",
				Health:          90,
				BuildNumber:     1,
				BuildRevision:   "abcdef1",
				BuildDate:       time.Unix(1483228800, 0),
				BuildResult:     "success",
				BuildDurationMs: 5,
				BuildArtifacts:  1,
				TestsDurationMs: 2,
				TestsSuites:     1,
				Tests:           3,
				TestsPassed:     2,
				TestsFailed:     1,
				TestsSkipped:    0,
			},
			"with only build data": {
				EntityName:      "bar",
				Health:          90,
				BuildNumber:     1,
				BuildRevision:   "abcdef1",
				BuildDate:       time.Unix(1483228800, 0),
				BuildResult:     "success",
				BuildDurationMs: 5,
				BuildArtifacts:  1,
			},
			"with no data": {
				EntityName: "baz",
			},
		}
		for name, ex := range expected {
			g.It("should return statistics about a job "+name, func() {
				job, err := fakeJenkins.GetJob(ex.EntityName)
				res := getJobStats(*job)
				g.Assert(err).Equal(nil)
				g.Assert(reflect.DeepEqual(res, ex)).Equal(true)
			})
		}
	})
}

func TestGetAllJobStats(t *testing.T) {
	g := goblin.Goblin(t)
	fakeJenkins := fakeJenkins()
	g.Describe("jenkins getAllJobStats()", func() {
		expected := []JobMetric{
			{
				EntityName:      "foo",
				Health:          90,
				BuildNumber:     1,
				BuildRevision:   "abcdef1",
				BuildDate:       time.Unix(1483228800, 0),
				BuildResult:     "success",
				BuildDurationMs: 5,
				BuildArtifacts:  1,
				TestsDurationMs: 2,
				TestsSuites:     1,
				Tests:           3,
				TestsPassed:     2,
				TestsFailed:     1,
				TestsSkipped:    0,
			},
			{
				EntityName:      "bar",
				Health:          90,
				BuildNumber:     1,
				BuildRevision:   "abcdef1",
				BuildDate:       time.Unix(1483228800, 0),
				BuildResult:     "success",
				BuildDurationMs: 5,
				BuildArtifacts:  1,
			},
			{
				EntityName: "baz",
			},
			{
				EntityName:      "baz/qux",
				Health:          90,
				BuildNumber:     1,
				BuildRevision:   "abcdef1",
				BuildDate:       time.Unix(1483228800, 0),
				BuildResult:     "success",
				BuildDurationMs: 5,
				BuildArtifacts:  1,
			},
		}
		res, err := getAllJobStats(fakeLog, fakeJenkins)
		g.It("should return statistics about many jobs", func() {
			g.Assert(err).Equal(nil)
			g.Assert(len(res)).Equal(len(expected))
			g.Assert(reflect.DeepEqual(res, expected)).Equal(true)
		})
	})
}

func TestGetNodeStats(t *testing.T) {
	g := goblin.Goblin(t)
	fakeJenkins := fakeJenkins()
	g.Describe("jenkins getNodeStats()", func() {
		expected := map[string]NodeMetric{
			"with two executors and is idle": {
				EntityName: "test-0",
				Online:     true,
				Idle:       true,
				Executors:  2,
			},
			"with four executors and not idle": {
				EntityName: "test-1",
				Online:     true,
				Idle:       false,
				Executors:  4,
			},
			"that is not online": {
				EntityName: "test-2",
				Online:     false,
				Idle:       true,
				Executors:  0,
			},
		}
		for name, ex := range expected {
			g.It("should return statistics about a node "+name, func() {
				node, err := fakeJenkins.GetNode(ex.EntityName)
				res := getNodeStats(*node)
				g.Assert(err).Equal(nil)
				g.Assert(reflect.DeepEqual(res, ex)).Equal(true)
			})
		}
	})
}

func TestGetAllNodeStats(t *testing.T) {
	g := goblin.Goblin(t)
	fakeJenkins := fakeJenkins()
	g.Describe("jenkins getAllNodeStats()", func() {
		expected := []NodeMetric{
			{
				EntityName: "test-0",
				Online:     true,
				Idle:       true,
				Executors:  2,
			},
			{
				EntityName: "test-1",
				Online:     true,
				Idle:       false,
				Executors:  4,
			},
			{
				EntityName: "test-2",
				Online:     false,
				Idle:       true,
				Executors:  0,
			},
		}
		g.It("should return statistics about many nodes", func() {
			res, err := getAllNodeStats(fakeLog, fakeJenkins)
			g.Assert(err).Equal(nil)
			g.Assert(reflect.DeepEqual(res, expected)).Equal(true)
		})
	})
}

func TestFindChildJobs(t *testing.T) {
	g := goblin.Goblin(t)
	fakeJenkins := fakeJenkins()
	g.Describe("jenkins findChildJobs()", func() {
		g.It("should recursively find child jobs", func() {
			parent, parentErr := fakeJenkins.GetJob("baz")
			g.Assert(parentErr).Equal(nil)

			expectedInner, expectedInnerErr := parent.GetInnerJob("qux")
			g.Assert(expectedInnerErr).Equal(nil)
			expected := []*gojenkins.Job{expectedInner}

			res, err := findChildJobs(fakeJenkins, parent)
			g.Assert(err).Equal(nil)
			g.Assert(reflect.DeepEqual(res, expected)).Equal(true)
		})
	})
}

func fakeJenkins() *gojenkins.Jenkins {
	jenkins := gojenkins.CreateJenkins(
		fakeConfig.JenkinsHost,
		fakeConfig.JenkinsAPIUser,
		fakeConfig.JenkinsAPIKey,
	)

	fakeJenkinsTransport := httpmock.NewMockTransport()
	registerResponders(fakeJenkinsTransport)
	jenkins.Requester.Client = &http.Client{
		Transport: fakeJenkinsTransport,
	}

	jenkins, err := jenkins.Init()
	if err != nil {
		panic("Error mocking connection to Jenkins")
	}

	return jenkins
}

// hash map of HTTP requests to mock
func registerResponders(transport *httpmock.MockTransport) {
	responses := []struct {
		Method   string
		Endpoint string
		Code     int
		Response string
	}{
		{"GET", "/", 200, `{"jobs":[{"name":"foo"},{"name":"bar"},{"name":"baz"}]}`},

		{"GET", "/job/foo", 200, `{"name":"foo","builds":[{"number":1}],"lastBuild":{"number":1},"healthReport":[{"score":100},{"score":80}],"previousBuild":{"number":1}}`},
		{"GET", "/job/bar", 200, `{"name":"bar","builds":[{"number":1}],"lastBuild":{"number":1},"healthReport":[{"score":100},{"score":80}],"previousBuild":{"number":1}}`},
		{"GET", "/job/baz", 200, `{"name":"baz","jobs":[{"name":"qux"}]}`},
		{"GET", "/job/baz/job/qux", 200, `{"name":"qux","builds":[{"number":1}],"lastBuild":{"number":1},"healthReport":[{"score":100},{"score":80}],"previousBuild":{"number":1}}`},

		{"GET", "/job/foo/1", 200, `{"id":"1","number":1,"timestamp":1483228800000,"duration":5,"result":"SUCCESS","actions":[{"_class":"hudson.plugins.git.util.BuildData","lastBuiltRevision":{"SHA1":"abcdef1"}}],"changeSet":{"kind":"git","items":[{}]},"artifacts":[{}]}`},
		{"GET", "/job/bar/1", 200, `{"id":"1","number":1,"timestamp":1483228800000,"duration":5,"result":"SUCCESS","actions":[{"_class":"hudson.plugins.git.util.BuildData","lastBuiltRevision":{"SHA1":"abcdef1"}}],"changeSet":{"kind":"git","items":[{}]},"artifacts":[{}]}`},
		{"GET", "/job/baz/job/qux/1", 200, `{"id":"1","number":1,"timestamp":1483228800000,"duration":5,"result":"SUCCESS","actions":[{"_class":"hudson.plugins.git.util.BuildData","lastBuiltRevision":{"SHA1":"abcdef1"}}],"changeSet":{"kind":"git","items":[{}]},"artifacts":[{}]}`},

		{"GET", "/job/foo/1/testReport", 200, `{"duration":2,"empty":false,"passCount":2,"failCount":1,"skipCount":0,"suites":[{"cases":[{},{},{}],"duration":2,"name":"test","id":null}]}`},
		{"GET", "/job/bar/1/testReport", 404, `{}`},
		{"GET", "/job/baz/job/qux/1/testReport", 404, `{}`},

		{"GET", "/computer", 200, `{"computer":[{"displayName":"test-0","executors":[{},{}],"idle":true,"offline":false},{"displayName":"test-1","executors":[{},{},{},{}],"idle":false,"offline":false},{"displayName":"test-2","executors":[],"idle":true,"offline":true}]}`},
		{"GET", "/computer/test-0", 200, `{"displayName":"test-0","executors":[{},{}],"idle":true,"offline":false}`},
		{"GET", "/computer/test-1", 200, `{"displayName":"test-1","executors":[{},{},{},{}],"idle":false,"offline":false}`},
		{"GET", "/computer/test-2", 200, `{"displayName":"test-2","executors":[],"idle":true,"offline":true}`},
	}

	extraslash := regexp.MustCompile("([^:])//+")
	for r := range responses {
		match := responses[r]
		url := extraslash.ReplaceAllString(strings.Join([]string{fakeConfig.JenkinsHost, match.Endpoint, "api", "json"}, "/"), "$1/")
		transport.RegisterResponder(match.Method, url, func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(match.Code, match.Response)
			resp.Header.Add("Content-Type", "application/json")
			resp.Header.Add("X-Jenkins", "mock")
			if testing.Verbose() {
				fmt.Println("httpmock: match", req.Method, req.URL, "->", match.Endpoint)
			}
			return resp, nil
		})
	}

	transport.RegisterNoResponder(func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(501, "{}")
		resp.Header.Add("X-Jenkins", "mock")
		if testing.Verbose() {
			fmt.Println("httpmock: no match", req.Method, req.URL)
		}
		return resp, nil
	})
}
