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
  "github.com/Sirupsen/logrus"
  "github.com/franela/goblin"
  "github.com/jarcoal/httpmock"
)

var (
  fakeLog           = logrus.New()
  fakeJenkinsConfig = JenkinsConfig{
    JenkinsHost:    "http://jenkins.mock",
    JenkinsAPIUser: "test-user",
    JenkinsAPIKey:  "test-pw",
  }
)

func TestValidateConfig(t *testing.T) {
  g := goblin.Goblin(t)
  g.Describe("jenkins validateConfig()", func() {
    g.It("should return an error when JenkinsHost is not set", func() {
      err := validateConfig(fakeLog, JenkinsConfig{})
      g.Assert(err == nil).IsFalse()
    })
    g.It("should return nil when JenkinsHost is set and JenkinsAPIUser and JenkinsAPIKey are not", func() {
      err := validateConfig(fakeLog, JenkinsConfig{
        JenkinsHost: "foo",
      })
      g.Assert(err).Equal(nil)
    })
    g.It("should return an error when JenkinsAPIUser is set but JenkinsAPIKey is not", func() {
      err := validateConfig(fakeLog, JenkinsConfig{
        JenkinsHost: "foo",
        JenkinsAPIUser: "bar",
      })
      g.Assert(err == nil).IsFalse()
    })
    g.It("should return an error when JenkinsAPIKey is set but JenkinsAPIUser is not", func() {
      err := validateConfig(fakeLog, JenkinsConfig{
        JenkinsHost: "foo",
        JenkinsAPIKey: "bar",
      })
      g.Assert(err == nil).IsFalse()
    })
    g.It("should return nil when all of the keys are set", func() {
      err := validateConfig(fakeLog, JenkinsConfig{
        JenkinsHost: "foo",
        JenkinsAPIUser: "bar",
        JenkinsAPIKey: "baz",
      })
      g.Assert(err).Equal(nil)
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
      g.Assert(len(res) > 0).IsTrue()
    })
    g.It("should have 'event_type' keys on everything", func() {
      for _, metric := range res {
        g.Assert(metric["event_type"] != nil).IsTrue()
      }
    })
    g.It("should have 'entity_name' keys on everything", func() {
      for _, metric := range res {
        g.Assert(metric["entity_name"] != nil).IsTrue()
      }
    })
    g.It("should have 'provider' keys on everything", func() {
      for _, metric := range res {
        g.Assert(metric["entity_name"] != nil).IsTrue()
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
        BuildResult: "success",
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
        BuildResult: "success",
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
        g.Assert(reflect.DeepEqual(res, ex)).IsTrue()
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
        BuildResult: "success",
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
        BuildResult: "success",
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
        BuildResult: "success",
        BuildDurationSecond: 5,
        BuildArtifacts: 1,
      },
    }
    g.It("should return statistics about many Jobs", func() {
      res, err := getAllJobStats(fakeLog, fakeJenkins)
      for i := range expected {
        g.Assert(err).Equal(nil)
        g.Assert(len(res)).Equal(len(expected))
        g.Assert(reflect.DeepEqual(res[i], expected[i])).IsTrue()
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
        g.Assert(reflect.DeepEqual(res, ex)).IsTrue()
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
      g.Assert(reflect.DeepEqual(res, expected)).IsTrue()
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
      g.Assert(reflect.DeepEqual(res, expected)).IsTrue()
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
    panic("Error mocking connection to Jenkins")
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
    { "GET", "/", 200, `{"jobs":[{"name":"foo"},{"name":"bar"},{"name":"baz"}]}` },

    { "GET", "/job/foo", 200, `{"name":"foo","builds":[{"number":1}],"lastBuild":{"number":1},"healthReport":[{"score":100},{"score":80}],"previousBuild":{"number":1}}` },
    { "GET", "/job/bar", 200, `{"name":"bar","builds":[{"number":1}],"lastBuild":{"number":1},"healthReport":[{"score":100},{"score":80}],"previousBuild":{"number":1}}` },
    { "GET", "/job/baz", 200, `{"name":"baz","jobs":[{"name":"qux"}]}` },
    { "GET", "/job/baz/job/qux", 200, `{"name":"qux","builds":[{"number":1}],"lastBuild":{"number":1},"healthReport":[{"score":100},{"score":80}],"previousBuild":{"number":1}}` },

    { "GET", "/job/foo/1", 200, `{"id":"1","number":1,"timestamp":1483228800000,"duration":5,"result":"SUCCESS","actions":[{"_class":"hudson.plugins.git.util.BuildData","lastBuiltRevision":{"SHA1":"abcdef1"}}],"changeSet":{"kind":"git","items":[{}]},"artifacts":[{}]}` },
    { "GET", "/job/bar/1", 200, `{"id":"1","number":1,"timestamp":1483228800000,"duration":5,"result":"SUCCESS","actions":[{"_class":"hudson.plugins.git.util.BuildData","lastBuiltRevision":{"SHA1":"abcdef1"}}],"changeSet":{"kind":"git","items":[{}]},"artifacts":[{}]}` },
    { "GET", "/job/baz/job/qux/1", 200, `{"id":"1","number":1,"timestamp":1483228800000,"duration":5,"result":"SUCCESS","actions":[{"_class":"hudson.plugins.git.util.BuildData","lastBuiltRevision":{"SHA1":"abcdef1"}}],"changeSet":{"kind":"git","items":[{}]},"artifacts":[{}]}` },

    { "GET", "/job/foo/1/testReport", 200, `{"duration":2,"empty":false,"passCount":2,"failCount":1,"skipCount":0,"suites":[{"cases":[{},{},{}],"duration":2,"name":"test","id":null}]}` },
    { "GET", "/job/bar/1/testReport", 404, `{}` },
    { "GET", "/job/baz/job/qux/1/testReport", 404, `{}` },

    { "GET", "/computer", 200, `{"computer":[{"displayName":"test-0","executors":[{},{}],"idle":true,"offline":false},{"displayName":"test-1","executors":[{},{},{},{}],"idle":false,"offline":false},{"displayName":"test-2","executors":[],"idle":true,"offline":true}]}` },
    { "GET", "/computer/test-0", 200, `{"displayName":"test-0","executors":[{},{}],"idle":true,"offline":false}` },
    { "GET", "/computer/test-1", 200, `{"displayName":"test-1","executors":[{},{},{},{}],"idle":false,"offline":false}` },
    { "GET", "/computer/test-2", 200, `{"displayName":"test-2","executors":[],"idle":true,"offline":true}` },
  }

  extraslash := regexp.MustCompile("([^:])//+")
  for _, match := range responses {
    url := extraslash.ReplaceAllString(strings.Join([]string{fakeJenkinsConfig.JenkinsHost, match.Endpoint, "api", "json"}, "/"), "$1/")
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
