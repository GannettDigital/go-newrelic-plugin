package jenkins

import (
  "net/http"
  "reflect"
  "regexp"
  "strings"
  "testing"
  "time"

  "github.com/bndr/gojenkins"
  "github.com/Sirupsen/logrus"
  "github.com/franela/goblin"
  "github.com/jarcoal/httpmock"
)

var (
  fakeLog           = logrus.New()
  fakeJenkinsConfig = JenkinsConfig{
    JenkinsHost:    "http://jenkins.mock",
    JenkinsAPIUser: "test",
    JenkinsAPIKey:  "test",
  }
)

func TestValidateConfig(t *testing.T) {
  g := goblin.Goblin(t)
  g.Describe("jenkins validateConfig()", func() {
    g.It("should return an error when JenkinsHost is not set", func() {
      e := validateConfig(fakeLog, JenkinsConfig{})
      g.Assert(e == nil).Equal(false)
    })
    g.It("should return nil when JenkinsHost is set and JenkinsAPIUser and JenkinsAPIKey are not", func() {
      e := validateConfig(fakeLog, JenkinsConfig{
        JenkinsHost: "foo",
      })
      g.Assert(e == nil).Equal(true)
    })
    g.It("should return an error when JenkinsAPIUser is set but JenkinsAPIKey is not", func() {
      e := validateConfig(fakeLog, JenkinsConfig{
        JenkinsHost: "foo",
        JenkinsAPIUser: "bar",
      })
      g.Assert(e == nil).Equal(false)
    })
    g.It("should return an error when JenkinsAPIKey is set but JenkinsAPIUser is not", func() {
      e := validateConfig(fakeLog, JenkinsConfig{
        JenkinsHost: "foo",
        JenkinsAPIKey: "bar",
      })
      g.Assert(e == nil).Equal(false)
    })
    g.It("should return nil when all of the keys are set", func() {
      e := validateConfig(fakeLog, JenkinsConfig{
        JenkinsHost: "foo",
        JenkinsAPIUser: "bar",
        JenkinsAPIKey: "baz",
      })
      g.Assert(e == nil).Equal(true)
    })
  })
}

func TestGetFullJobName(t *testing.T) {
  g := goblin.Goblin(t)
  g.Describe("jenkins getFullJobName()", func() {
    var tests = map[string]gojenkins.Job{
      "foo":         gojenkins.Job{Base: "/job/foo"},
      "foo/bar":     gojenkins.Job{Base: "/job/foo/job/bar"},
      "foo/bar/baz": gojenkins.Job{Base: "/job/foo/job/bar/job/baz"},
    }
    g.It("should return job name prefixed by parent job names", func() {
      for expected, job := range tests {
        res := getFullJobName(job)
        g.Assert(res).Equal(expected)
      }
    })
  })
}

func TestGetJenkins(t *testing.T) {
  g := goblin.Goblin(t)
  fakeJenkins := fakeJenkins()
  g.Describe("jenkins getJenkins()", func() {
    res := getJenkins(fakeJenkinsConfig)
    g.It("should connect to the right Jenkins", func() {
      g.Assert(res.Server).Equal("http://jenkins.mock")
    })
    g.It("should have authorization data", func() {
      g.Assert(res.Requester.BasicAuth.Username).Equal("test")
      g.Assert(res.Requester.BasicAuth.Password).Equal("test")
    })
    g.It("should be using httpmock.MockTransport for requests while testing", func() {
      t := reflect.TypeOf(fakeJenkins.Requester.Client.Transport)
      g.Assert(t.String()).Equal("*httpmock.MockTransport")
    })
  })
}

func TestGetMetrics(t *testing.T) {
  g := goblin.Goblin(t)
  fakeJenkins := fakeJenkins()
  g.Describe("jenkins getMetrics()", func() {
    res, err := getMetrics(fakeLog, fakeJenkins)
    g.It("should return metric data", func() {
      g.Assert(err == nil).Equal(true)
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
        g.Assert(metric["entity_name"] != nil).Equal(true)
      }
    })
  })
}

func TestGetJobStats(t *testing.T) {
  g := goblin.Goblin(t)
  fakeJenkins := fakeJenkins()
  g.Describe("jenkins getJobStats()", func() {
    expected := []JobMetric{
      {
        EntityName: "foo",
        Health: 90,
        BuildNumber: 1,
        BuildRevision: "abcdef1",
        BuildDate: time.Unix(1483228800, 0),
        BuildResult: "passed",
        BuildDurationSecond: 5,
        BuildArtifacts: 1,
        TestsDurationSecond: 2,
        TestsSuites: 1,
        Tests: 3,
        TestsPassed: 2,
        TestsFailed: 1,
        TestsSkipped: 0,
      },
      {
        EntityName: "bar",
        Health: 90,
        BuildNumber: 1,
        BuildRevision: "abcdef1",
        BuildDate: time.Unix(1483228800, 0),
        BuildResult: "passed",
        BuildDurationSecond: 5,
        BuildArtifacts: 1,
      },
      {
        EntityName: "baz",
      },
    }
    g.It("should return statistics about a Job object", func() {
      for _, ex := range expected {
        job, err := fakeJenkins.GetJob(ex.EntityName)
        res := getJobStats(*job)
        g.Assert(err).Equal(nil)
        g.Assert(reflect.DeepEqual(res, ex)).Equal(true)
      }
    })
  })
}

func TestGetAllJobStats(t *testing.T) {
  g := goblin.Goblin(t)
  fakeJenkins := fakeJenkins()
  g.Describe("jenkins getAllJobStats()", func() {
    expected := []JobMetric{
      {
        EntityName: "foo",
        Health: 90,
        BuildNumber: 1,
        BuildRevision: "abcdef1",
        BuildDate: time.Unix(1483228800, 0),
        BuildResult: "passed",
        BuildDurationSecond: 5,
        BuildArtifacts: 1,
        TestsDurationSecond: 2,
        TestsSuites: 1,
        Tests: 3,
        TestsPassed: 2,
        TestsFailed: 1,
        TestsSkipped: 0,
      },
      {
        EntityName: "bar",
        Health: 90,
        BuildNumber: 1,
        BuildRevision: "abcdef1",
        BuildDate: time.Unix(1483228800, 0),
        BuildResult: "passed",
        BuildDurationSecond: 5,
        BuildArtifacts: 1,
      },
      {
        EntityName: "baz",
      },
      {
        EntityName: "baz/qux",
        Health: 90,
        BuildNumber: 1,
        BuildRevision: "abcdef1",
        BuildDate: time.Unix(1483228800, 0),
        BuildResult: "passed",
        BuildDurationSecond: 5,
        BuildArtifacts: 1,
      },
    }
    g.It("should return statistics about many Jobs", func() {
      _, warmupErr := fakeJenkins.GetAllJobs()
      res, err := getAllJobStats(fakeLog, fakeJenkins)
      for i := range expected {
        g.Assert(warmupErr).Equal(nil)
        g.Assert(err).Equal(nil)
        g.Assert(len(res) == len(expected)).Equal(true)
        g.Assert(reflect.DeepEqual(res[i], expected[i])).Equal(true)
      }
    })
  })
}

func TestGetNodeStats(t *testing.T) {
  g := goblin.Goblin(t)
  fakeJenkins := fakeJenkins()
  g.Describe("jenkins getNodeStats()", func() {
    expected := []NodeMetric{
      {
        EntityName: "test-0",
        Online: true,
        Idle: true,
        Executors: 2,
      },
      {
        EntityName: "test-1",
        Online: true,
        Idle: false,
        Executors: 4,
      },
      {
        EntityName: "test-2",
        Online: false,
        Idle: true,
        Executors: 0,
      },
    }
    g.It("should return statistics about a Node object", func() {
      for _, ex := range expected {
        node, err := fakeJenkins.GetNode(ex.EntityName)
        res := getNodeStats(*node)
        g.Assert(err).Equal(nil)
        g.Assert(reflect.DeepEqual(res, ex)).Equal(true)
      }
    })
  })
}

func TestGetAllNodeStats(t *testing.T) {
  g := goblin.Goblin(t)
  fakeJenkins := fakeJenkins()
  g.Describe("jenkins getAllNodeStats()", func() {
    expected := []NodeMetric{
      {
        EntityName: "test-0",
        Online: true,
        Idle: true,
        Executors: 2,
      },
      {
        EntityName: "test-1",
        Online: true,
        Idle: false,
        Executors: 4,
      },
      {
        EntityName: "test-2",
        Online: false,
        Idle: true,
        Executors: 0,
      },
    }
    g.It("should return statistics about many Nodes", func() {
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
      job, jobErr := fakeJenkins.GetJob("baz")
      res, err := findChildJobs(fakeJenkins, job)
      g.Assert(jobErr).Equal(nil)
      g.Assert(err).Equal(nil)
      g.Assert(len(res)).Equal(1)
      g.Assert(res[0].GetName()).Equal("qux")
    })
  })
}

func fakeJenkins() *gojenkins.Jenkins {
  jenkins := gojenkins.CreateJenkins(
    fakeJenkinsConfig.JenkinsHost,
    fakeJenkinsConfig.JenkinsAPIUser,
    fakeJenkinsConfig.JenkinsAPIKey,
  )

  fakeJenkinsTransport := httpmock.NewMockTransport()
  registerResponders(fakeJenkinsTransport)
  jenkins.Requester.Client = &http.Client{
    Transport: fakeJenkinsTransport,
  }

  jenkins, err := jenkins.Init()
  if err != nil {
    fakeLog.WithError(err).Fatal("Error mocking connection to Jenkins")
  }

  return jenkins
}

// hash map of HTTP requests to mock
func registerResponders(transport *httpmock.MockTransport) {
  responses := []struct{
    Method   string
    Endpoint string
    Code     int
    Response string
  }{
    { "GET", "/", 200, `{"jobs":[{"name":"foo","url":"http://jenkins.mock/job/foo/"},{"name":"bar","url":"http://jenkins.mock/job/bar/"},{"name":"baz","url":"http://jenkins.mock/job/baz/"}]}` },

    { "GET", "/job/foo", 200, `{"name":"foo","displayName":"foo","builds":[{"number":1,"url":"http://jenkins.mock/job/foo/1/"}],"lastBuild":{"number":1,"url":"http://jenkins.mock/job/foo/1/"},"healthReport":[{"score":100},{"score":80}],"previousBuild":{"number":1,"url":"http://jenkins.mock/job/foo/1/"}}` },
    { "GET", "/job/bar", 200, `{"name":"bar","displayName":"bar","builds":[{"number":1,"url":"http://jenkins.mock/job/bar/1/"}],"lastBuild":{"number":1,"url":"http://jenkins.mock/job/bar/1/"},"healthReport":[{"score":100},{"score":80}],"previousBuild":{"number":1,"url":"http://jenkins.mock/job/bar/1/"}}` },
    { "GET", "/job/baz", 200, `{"name":"baz","displayName":"baz","jobs":[{"name":"qux","url":"http://jenkins.mock/job/baz/job/qux/","color":"blue"}]}` },
    { "GET", "/job/baz/job/qux", 200, `{"name":"qux","displayName":"qux","builds":[{"number":1,"url":"http://jenkins.mock/job/baz/job/qux/1/"}],"lastBuild":{"number":1,"url":"http://jenkins.mock/job/baz/job/qux/1/"},"healthReport":[{"score":100},{"score":80}],"previousBuild":{"number":1,"url":"http://jenkins.mock/job/baz/job/qux/1/"}}` },

    { "GET", "/job/foo/1", 200, `{"id":"1","number":1,"timestamp":1483228800000,"duration":5,"result":"PASSED","actions":[{"_class":"hudson.plugins.git.util.BuildData","lastBuiltRevision":{"SHA1":"abcdef1"}}],"changeSet":{"kind":"git","items":[{}]},"artifacts":[{}]}` },
    { "GET", "/job/bar/1", 200, `{"id":"1","number":1,"timestamp":1483228800000,"duration":5,"result":"PASSED","actions":[{"_class":"hudson.plugins.git.util.BuildData","lastBuiltRevision":{"SHA1":"abcdef1"}}],"changeSet":{"kind":"git","items":[{}]},"artifacts":[{}]}` },
    { "GET", "/job/baz/job/qux/1", 200, `{"id":"1","number":1,"timestamp":1483228800000,"duration":5,"result":"PASSED","actions":[{"_class":"hudson.plugins.git.util.BuildData","lastBuiltRevision":{"SHA1":"abcdef1"}}],"changeSet":{"kind":"git","items":[{}]},"artifacts":[{}]}` },

    { "GET", "/job/foo/1/testReport", 200, `{"duration":2,"empty":false,"passCount":2,"failCount":1,"skipCount":0,"suites":[{"cases":[{},{},{}],"duration":2,"name":"test","id":null}]}` },
    { "GET", "/job/bar/1/testReport", 404, `{}` },
    { "GET", "/job/baz/job/qux/1/testReport", 404, `{}` },

    { "GET", "/computer", 200, `{"displayName":"Nodes","busyExecutors":0,"totalExecutors":6,"computer":[{"displayName":"test-0","executors":[{},{}],"idle":true,"offline":false},{"displayName":"test-1","executors":[{},{},{},{}],"idle":false,"offline":false},{"displayName":"test-2","executors":[],"idle":true,"offline":true}]}` },
    { "GET", "/computer/test-0", 200, `{"displayName":"test-0","executors":[{},{}],"idle":true,"offline":false}` },
    { "GET", "/computer/test-1", 200, `{"displayName":"test-1","executors":[{},{},{},{}],"idle":false,"offline":false}` },
    { "GET", "/computer/test-2", 200, `{"displayName":"test-2","executors":[],"idle":true,"offline":true}` },
  }

  re := regexp.MustCompile("([^:])//+")
  for _, r := range responses {
    url := re.ReplaceAllString(strings.Join([]string{fakeJenkinsConfig.JenkinsHost, r.Endpoint, "api", "json"}, "/"), "$1/")
    transport.RegisterResponder(r.Method, url, httpmock.NewStringResponder(r.Code, r.Response))
  }

  transport.RegisterNoResponder(func(req *http.Request) (*http.Response, error) {
    response := httpmock.NewStringResponse(501, "{}")
    fakeLog.WithFields(logrus.Fields{
      "request": req,
      "response": response,
    }).Warn("Unmocked HTTP request")
    return response, nil
  })
}
