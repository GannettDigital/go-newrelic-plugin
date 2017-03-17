package jenkins

import (
  "bytes"
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
  fakeLog           *logrus.Logger
  fakeLogReporter   = new(bytes.Buffer)
  fakeJenkinsConfig JenkinsConfig
  fakeJenkins       *gojenkins.Jenkins
  fakeJenkinsErr    error
)

func init() {
  fakeLog = logrus.New()
  fakeLog.Level = logrus.PanicLevel
  fakeLog.Out = fakeLogReporter

  fakeJenkinsConfig = JenkinsConfig{
    JenkinsHost:    "http://jenkins.mock",
    JenkinsAPIUser: "test",
    JenkinsAPIKey:  "test",
  }

  fakeJenkins = gojenkins.CreateJenkins(
    fakeJenkinsConfig.JenkinsHost,
    fakeJenkinsConfig.JenkinsAPIUser,
    fakeJenkinsConfig.JenkinsAPIKey,
  )

  fakeJenkinsTransport := httpmock.NewMockTransport()
  registerResponders(fakeJenkinsTransport)
  fakeJenkins.Requester.Client = &http.Client{
    Transport: fakeJenkinsTransport,
  }

  fakeJenkins, fakeJenkinsErr = fakeJenkins.Init()
  if fakeJenkinsErr != nil {
    fakeLog.WithError(fakeJenkinsErr).Fatal("Error mocking connection to Jenkins")
  }
}

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

func TestGetMetrics(t *testing.T) {
  g := goblin.Goblin(t)
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

func TestGetJenkins(t *testing.T) {
  g := goblin.Goblin(t)
  g.Describe("jenkins getJenkins()", func() {
    res := getJenkins(fakeJenkinsConfig)
    g.It("should connect to the right Jenkins", func() {
      g.Assert(res.Server).Equal("http://jenkins.mock")
    })
    g.It("should have authorization data", func() {
      g.Assert(res.Requester.BasicAuth.Username).Equal("test")
      g.Assert(res.Requester.BasicAuth.Password).Equal("test")
    })
  })
}

func TestGetAllJobStats(t *testing.T) {
  g := goblin.Goblin(t)
  g.Describe("jenkins getAllJobStats()", func() {
    g.It("should return statistics about many Jobs", func() {
      res, err := getAllJobStats(fakeLog, fakeJenkins)
      g.Assert(err).Equal(nil)
      g.Assert(len(res) == 4).Equal(true)
    })
  })
}

func TestFindChildJobs(t *testing.T) {
  g := goblin.Goblin(t)
  g.Describe("jenkins findChildJobs()", func() {
    g.It("should recursively find child jobs", func() {
      job, jobErr := fakeJenkins.GetJob("baz")
      res, err := findChildJobs(fakeJenkins, job)
      g.Assert(jobErr).Equal(nil)
      g.Assert(err).Equal(nil)
      g.Assert(len(res) == 1).Equal(true)
    })
  })
}

func TestGetJobStats(t *testing.T) {
  g := goblin.Goblin(t)
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

func TestGetAllNodesStats(t *testing.T) {
  g := goblin.Goblin(t)
  g.Describe("jenkins getAllNodeStats()", func() {
    res, err := getAllNodeStats(fakeLog, fakeJenkins)
    g.It("should return statistics about many Nodes", func() {
      g.Assert(err).Equal(nil)
      g.Assert(len(res) == 3).Equal(true)
    })
  })
}

func TestGetNodeStats(t *testing.T) {
  g := goblin.Goblin(t)
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

func TestGetFullJobName(t *testing.T) {
  g := goblin.Goblin(t)
  g.Describe("jenkins getFullJobName()", func() {
    var tests = []struct{
      Job      gojenkins.Job
      FullName string
    }{
      {
        Job:      gojenkins.Job{Base: "/job/foo"},
        FullName: "foo",
      },
      {
        Job:      gojenkins.Job{Base: "/job/foo/job/bar"},
        FullName: "foo/bar",
      },
      {
        Job:      gojenkins.Job{Base: "/job/foo/job/bar/job/baz"},
        FullName: "foo/bar/baz",
      },
    }
    g.It("should return full job name", func() {
      for _, test := range tests {
        res := getFullJobName(test.Job)
        g.Assert(res).Equal(test.FullName)
      }
    })
  })
}

// hash map of HTTP requests to mock
func registerResponders(transport *httpmock.MockTransport) {
  responses := map[string]string{
    "/": "{\"jobs\":[{\"name\":\"foo\",\"url\":\"http://jenkins.mock/job/foo/\"},{\"name\":\"bar\",\"url\":\"http://jenkins.mock/job/bar/\"},{\"name\":\"baz\",\"url\":\"http://jenkins.mock/job/baz/\"}]}",

    "/job/foo": "{\"name\":\"foo\",\"displayName\":\"foo\",\"url\":\"http://jenkins.mock/job/foo/\",\"builds\":[{\"number\":1,\"url\":\"http://jenkins.mock/job/foo/1/\"}],\"lastBuild\":{\"number\":1,\"url\":\"http://jenkins.mock/job/foo/1/\"},\"healthReport\":[{\"score\":100},{\"score\":80}],\"previousBuild\":{\"number\":1,\"url\":\"http://jenkins.mock/job/foo/1/\"}}",
    "/job/bar": "{\"name\":\"bar\",\"displayName\":\"bar\",\"url\":\"http://jenkins.mock/job/bar/\",\"builds\":[{\"number\":1,\"url\":\"http://jenkins.mock/job/bar/1/\"}],\"lastBuild\":{\"number\":1,\"url\":\"http://jenkins.mock/job/bar/1/\"},\"healthReport\":[{\"score\":100},{\"score\":80}],\"previousBuild\":{\"number\":1,\"url\":\"http://jenkins.mock/job/bar/1/\"}}",
    "/job/baz": "{\"name\":\"baz\",\"displayName\":\"baz\",\"url\":\"http://jenkins.mock/job/baz/\",\"jobs\":[{\"name\":\"qux\",\"url\":\"http://jenkins.mock/job/baz/job/qux/\",\"color\":\"blue\"}],\"healthReport\":[{\"score\":100}]}",
    "/job/baz/job/qux": "{\"name\":\"qux\",\"displayName\":\"qux\",\"url\":\"http://jenkins.mock/job/baz/job/qux/\",\"builds\":[{\"number\":1,\"url\":\"http://jenkins.mock/job/baz/job/qux/1/\"}],\"lastBuild\":{\"number\":1,\"url\":\"http://jenkins.mock/job/baz/job/qux/1/\"},\"healthReport\":[{\"score\":100},{\"score\":80}],\"previousBuild\":{\"number\":1,\"url\":\"http://jenkins.mock/job/baz/job/qux/1/\"}}",

    "/job/foo/1": "{\"id\":\"1\",\"number\":1,\"timestamp\":1483228800000,\"url\":\"http://jenkins.mock/job/foo/1/\",\"duration\":5,\"result\":\"PASSED\",\"actions\":[{\"_class\":\"hudson.plugins.git.util.BuildData\",\"lastBuiltRevision\":{\"SHA1\":\"abcdef1\"}}],\"changeSet\":{\"kind\":\"git\",\"items\":[{}]},\"artifacts\":[{}]}",
    "/job/bar/1": "{\"id\":\"1\",\"number\":1,\"timestamp\":1483228800000,\"url\":\"http://jenkins.mock/job/bar/1/\",\"duration\":5,\"result\":\"PASSED\",\"actions\":[{\"_class\":\"hudson.plugins.git.util.BuildData\",\"lastBuiltRevision\":{\"SHA1\":\"abcdef1\"}}],\"changeSet\":{\"kind\":\"git\",\"items\":[{}]},\"artifacts\":[{}]}",
    "/job/baz/job/qux/1": "{\"id\":\"1\",\"number\":1,\"timestamp\":1483228800000,\"url\":\"http://jenkins.mock/job/baz/1/\",\"duration\":5,\"result\":\"PASSED\",\"actions\":[{\"_class\":\"hudson.plugins.git.util.BuildData\",\"lastBuiltRevision\":{\"SHA1\":\"abcdef1\"}}],\"changeSet\":{\"kind\":\"git\",\"items\":[{}]},\"artifacts\":[{}]}",

    "/job/foo/1/testReport": "{\"duration\":2,\"empty\":false,\"passCount\":2,\"failCount\":1,\"skipCount\":0,\"suites\":[{\"cases\":[{},{},{}],\"duration\":2,\"name\":\"test\",\"id\":null}]}",

    "/computer": "{\"displayName\":\"nodes\",\"busyExecutors\":0,\"totalExecutors\":6,\"computer\":[{\"displayName\":\"test-0\",\"executors\":[{},{}],\"idle\":true,\"offline\":false},{\"displayName\":\"test-1\",\"executors\":[{},{},{},{}],\"idle\":false,\"offline\":false},{\"displayName\":\"test-2\",\"executors\":[],\"idle\":true,\"offline\":true}]}",
    "/computer/test-0": "{\"displayName\":\"test-0\",\"executors\":[{},{}],\"idle\":true,\"offline\":false}",
    "/computer/test-1": "{\"displayName\":\"test-1\",\"executors\":[{},{},{},{}],\"idle\":false,\"offline\":false}",
    "/computer/test-2": "{\"displayName\":\"test-2\",\"executors\":[],\"idle\":true,\"offline\":true}",
  }

  re := regexp.MustCompile("([^:])//+")
  for endpoint, response := range responses {
    url := re.ReplaceAllString(strings.Join([]string{fakeJenkinsConfig.JenkinsHost, endpoint, "api", "json"}, "/"), "$1/")
    transport.RegisterResponder("GET", url, httpmock.NewStringResponder(200, response))
  }

  transport.RegisterNoResponder(func(req *http.Request) (*http.Response, error) {
    fakeLog.WithField("request", req).Warn("Unmocked HTTP request")
    response := httpmock.NewStringResponse(501, "[]")
    return response, nil
  })
}
