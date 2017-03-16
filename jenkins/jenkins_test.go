package jenkins

import (
  "encoding/json"
  "fmt"
  // "net/http"
  "reflect"
  "testing"

  // fake "github.com/GannettDigital/paas-api-utils/utilsHTTP/fake"

  "github.com/bndr/gojenkins"
  "github.com/Sirupsen/logrus"
  "github.com/franela/goblin"
)

var fakeJenkinsConfig JenkinsConfig

func init() {
  fakeJenkinsConfig = JenkinsConfig{
    JenkinsHost:     "http://jenkins.test",
    JenkinsAPIUser: "test",
    JenkinsAPIKey:  "test",
  }
}

func TestJenkins(t *testing.T) {
  g := goblin.Goblin(t)


  g.Describe("jenkins", func() {
    var (
      fakeLog     *logrus.Logger
      fakeJenkins gojenkins.Jenkins
      fakeJobs    []gojenkins.Job
      fakeNodes   []gojenkins.Node
    )

    g.Before(func() {
      fakeJenkins = gojenkins.Jenkins{}
      fakeJobs = mockJobs()
      fakeNodes = mockNodes()
    })

    // core collector things
    g.Describe("JenkinsCollector()", func() {

    })
    g.Describe("getJenkins()", func() {

    })
    g.Describe("getJenkinsData()", func() {

    })

    // job things
    g.Describe("getAllJobStats()", func() {

    })
    g.Describe("findChildJobs()", func() {

    })
    g.Describe("getJobStats()", func() {
      expected := []map[string]interface{}{
        {
          "jenkins.job.name": "some-job",
          "jenkins.job.health": 90,
          "jenkins.job.build.number": 1,
          "jenkins.job.build.revision": "abcdef1",
          "jenkins.job.build.date": 14832288000000,
          "jenkins.job.build.result": "passed",
          "jenkins.job.build.duration": 5,
          "jenkins.job.build.artifacts": 1,
          "jenkins.job.tests.duration": 2,
          "jenkins.job.tests.suites": 1,
          "jenkins.job.tests.total": 3,
          "jenkins.job.tests.passed": 2,
          "jenkins.job.tests.failed": 1,
          "jenkins.job.tests.skipped": 0,
        },
        {
          "jenkins.node.name": "job-without-tests",
          "jenkins.job.health": 90,
          "jenkins.job.build.number": 1,
          "jenkins.job.build.revision": "abcdef1",
          "jenkins.job.build.date": 14832288000000,
          "jenkins.job.build.result": "passed",
          "jenkins.job.build.duration": 5,
          "jenkins.job.build.artifacts": 1,
        },
        {
          "jenkins.job.name": "job-folder",
        },
      }

      for i, test := range fakeJobs {
        g.It("should return statistics about a Job object", func() {
          res := getJobStats(test)
          g.Assert(reflect.DeepEqual(res, expected[i])).Equal(true)
        })
      }
    })

    // node things
    g.Describe("getAllNodeStats()", func() {
      expected := []map[string]interface{}{
        {
          "jenkins.node.name": "test-node-0",
          "jenkins.node.online": true,
          "jenkins.node.idle": true,
          "jenkins.node.executors": 2,
        },
        {
          "jenkins.node.name": "test-node-1",
          "jenkins.node.online": true,
          "jenkins.node.idle": false,
          "jenkins.node.executors": 4,
        },
        {
          "jenkins.node.name": "test-node-2",
          "jenkins.node.online": false,
          "jenkins.node.idle": true,
          "jenkins.node.executors": 0,
        },
      }
      g.It("should return statistics about many Nodes", func() {
        res, err := getAllNodeStats(&fakeJenkins)
        g.Assert(err).Equal(nil)
        g.Assert(reflect.DeepEqual(res, expected)).Equal(true)
      })
    })
    g.Describe("getNodeStats()", func() {
      fakeNodes := mockNodes()
      expected := []map[string]interface{}{
        {
          "jenkins.node.name": "test-node-0",
          "jenkins.node.online": true,
          "jenkins.node.idle": true,
          "jenkins.node.executors": 2,
        },
        {
          "jenkins.node.name": "test-node-1",
          "jenkins.node.online": true,
          "jenkins.node.idle": false,
          "jenkins.node.executors": 4,
        },
        {
          "jenkins.node.name": "test-node-2",
          "jenkins.node.online": false,
          "jenkins.node.idle": true,
          "jenkins.node.executors": 0,
        },
      }
      for i := range expected {
        g.It("should return statistics about a Node object", func() {
          res := getNodeStats(fakeNodes[i])
          g.Assert(reflect.DeepEqual(res, expected[i])).Equal(true)
        })
      }
    })

    // helpers
    g.Describe("getFullJobName()", func() {
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
      for _, test := range tests {
        g.It("should return full job name", func() {
          res := getFullJobName(test.Job)
          g.Assert(res).Equal(test.FullName)
        })
      }
    })
    g.Describe("mergeMaps()", func() {
      mapA, mapB, expected := map[string]interface{}{
        "thing1": "stuff",
        "thing2": "morestuff",
      }, map[string]interface{}{
        "thing2": "otherstuff",
        "thing3": "alsostuff",
      }, map[string]interface{}{
        "thing1": "stuff",
        "thing2": "otherstuff",
        "thing3": "alsostuff",
      }
      g.It("merges two maps into one map", func() {
        res := mergeMaps(mapA, mapB)
        g.Assert(reflect.DeepEqual(res, expected)).Equal(true)
      })
    })
  })
}

func mockJobs() (fakeJobs []gojenkins.Job) {
  /* fakeJobResponsesJSON := [][]byte{
    []byte("{\"name\":\"some-job\",\"displayName\":\"some-job\",\"url\":\"http://jenkins.test/job/some-job/\",\"healthReport\":[{\"score\":80},{\"score\":100}],\"lastBuild\":{\"number\":1,\"url\":\"http://jenkins.test/job/some-job/1/\"}}"),
    []byte("{\"name\":\"job-without-tests\",\"displayName\":\"job-without-tests\",\"url\":\"http://jenkins.test/job/job-without-tests/\",\"healthReport\":[{\"score\":80},{\"score\":100}],\"lastBuild\":{\"number\":1,\"\":\"http://jenkins.test/job/job-without-tests/1/\"}}"),
    []byte("{\"name\":\"job-folder\",\"displayName\":\"job-folder\",\"url\":\"http://jenkins.test/job/job-folder/\",\"healthReport\":[],\"jobs\":[{\"name\":\"sub-job-1\",\"url\":\"http://jenkins.test/job/job-folder/job/sub-job-1/\"}]}"),
  }
  fakeBuildResponsesResultJSON := [][]byte{
    []byte("{\"artifacts\":[{}],\"number\":1,\"id\":\"1\",\"result\":\"SUCCESS\",\"timestamp\":14832288000000,\"duration\":5,\"url\":\"http://jenkins.test/job/some-job/1/\",\"changeSet\":{\"kind\":\"git\",\"items\":[{\"commitId\":\"abcdef1\"}]}}"),
    []byte("{\"artifacts\":[{}],\"number\":1,\"id\":\"1\",\"result\":\"SUCCESS\",\"timestamp\":14832288000000,\"duration\":5,\"url\":\"http://jenkins.test/job/some-job/1/\",\"changeSet\":{\"kind\":\"git\",\"items\":[{\"commitId\":\"abcdef1\"}]}}"),
    nil,
  }
  fakeTestResultJSON := [][]byte{
    []byte("{\"duration\":2,\"passCount\":2,\"failCount\":1,\"skipCount\":0,\"suites\":[{\"cases\":[{}, {}, {}]}]}"),
    nil,
    nil,
  }

  for i := range fakeJobResponsesJSON {
    var fakeJobResponse gojenkins.JobResponse
    var fakeJob gojenkins.Job

    json.Unmarshal(fakeJobResponsesJSON[i], &fakeJobResponse)
    fakeJob = gojenkins.Job{Raw: &fakeJobResponse}
    fakeJobs = append(fakeJobs, fakeJob)
  } */

  return fakeJobs
}

func mockNodes() (fakeNodes []gojenkins.Node) {
  fakeNodeResponses := [][]byte{
    []byte("{\"displayName\":\"test-node-0\",\"executors\":[{},{}],\"idle\":true,\"offline\":false,\"numExecutors\":2}"),
    []byte("{\"displayName\":\"test-node-1\",\"executors\":[{},{},{},{}],\"idle\":false,\"offline\":false,\"numExecutors\":4}"),
    []byte("{\"displayName\":\"test-node-2\",\"executors\":[],\"idle\":true,\"offline\":true,\"numExecutors\":0}"),
  }

  for i := range fakeNodeResponses {
    var fakeNodeResponse gojenkins.NodeResponse
    json.Unmarshal(fakeNodeResponses[i], &fakeNodeResponse)
    fakeNode := gojenkins.Node{Raw: &fakeNodeResponse, Base:fmt.Sprintf("/computer/test-node-%v", string(i))}
    fakeNodes = append(fakeNodes, fakeNode)
  }

  return fakeNodes
}
