package jenkins

import (
  "encoding/json"
  "fmt"
  "os"
  "strings"

  "github.com/bndr/gojenkins"
  "github.com/Sirupsen/logrus"
)

const NAME string = "jenkins"
const PROVIDER string = "jenkins"
const PROTOCOL_VERSION string = "1"

// JenkinsConfig is the keeper of the config
type JenkinsConfig struct {
  JenkinsAPIUser string
  JenkinsAPIKey  string
  JenkinsHost    string
}

// InventoryData is the data type for inventory data produced by a plugin data
// source and emitted to the agent's inventory data store
type InventoryData map[string]interface{}

// MetricData is the data type for events produced by a plugin data source and
// emitted to the agent's metrics data store
type MetricData map[string]interface{}

// EventData is the data type for single shot events
type EventData map[string]interface{}

// PluginData defines the format of the output JSON that plugins will return
type PluginData struct {
  Name            string                   `json:"name"`
  ProtocolVersion string                   `json:"protocol_version"`
  PluginVersion   string                   `json:"plugin_version"`
  Metrics         []MetricData             `json:"metrics"`
  Inventory       map[string]InventoryData `json:"inventory"`
  Events          []EventData              `json:"events"`
  Status          string                   `json:"status"`
}

// OutputJSON takes an object and prints it as a JSON string to the stdout.
// If the pretty attribute is set to true, the JSON will be idented for easy reading.
func OutputJSON(data interface{}, pretty bool) error {
  var output []byte
  var err error

  if pretty {
    output, err = json.MarshalIndent(data, "", "\t")
  } else {
    output, err = json.Marshal(data)
  }

  if err != nil {
    return fmt.Errorf("Error outputting JSON: %s", err)
  }

  if string(output) == "null" {
    fmt.Println("[]")
  } else {
    fmt.Println(string(output))
  }

  return nil
}

func Run(log *logrus.Logger, prettyPrint bool, version string) {

  // Initialize the output structure
  var data = PluginData{
    Name:            NAME,
    ProtocolVersion: PROTOCOL_VERSION,
    PluginVersion:   version,
    Inventory:       make(map[string]InventoryData),
    Metrics:         make([]MetricData, 0),
    Events:          make([]EventData, 0),
    Status:          "ok",
  }

  // get config from env vars
  var config = JenkinsConfig{
    JenkinsHost:     os.Getenv("JENKINS_HOST"),
    JenkinsAPIUser:  os.Getenv("JENKINS_API_USER"),
    JenkinsAPIKey:   os.Getenv("JENKINS_API_KEY"),
  }
  validateConfig(log, config)

  jenkins, jenkinsErr := getJenkins(config).Init()
  if jenkinsErr != nil {
    log.WithError(jenkinsErr).Fatal("Error connecting to Jenkins")
  }

  metrics, metricsErr := getMetrics(log, jenkins)
  fatalIfErr(log, metricsErr)

  data.Metrics = append(data.Metrics, metrics...)
  fatalIfErr(log, OutputJSON(data, prettyPrint))
}

func validateConfig(log *logrus.Logger, config JenkinsConfig) {
  if config.JenkinsHost == "" {
    log.Fatal("config: JENKINS_HOST must be set")
  }
  if config.JenkinsAPIUser != "" && config.JenkinsAPIKey == "" {
    log.Fatal("config: you must also set JENKINS_API_KEY when JENKINS_API_USER is set")
  }
  if config.JenkinsAPIUser == "" && config.JenkinsAPIKey != "" {
    log.Fatal("config: you must also set JENKINS_API_USER when JENKINS_API_KEY is set")
  }
}

func fatalIfErr(log *logrus.Logger, err error) {
  if err != nil {
    log.WithError(err).Fatal("can't continue")
  }
}

func getMetrics(log *logrus.Logger, jenkins *gojenkins.Jenkins) (records []MetricData, err error) {
  nodeData, err := getAllNodeStats(log, jenkins)
  if err != nil {
    return nil, err
  }
  for _, node := range nodeData {
    record := mergeMaps(map[string]interface{}{
      "event_type": "LoadBalancerSample",
      "provider": "jenkins.node",
    }, node)
    records = append(records, record)
  }

  jobData, err := getAllJobStats(log, jenkins)
  if err != nil {
    return nil, err
  }
  for _, job := range jobData {
    record := mergeMaps(map[string]interface{}{
      "event_type": "DatastoreSample",
      "provider": "jenkins.job",
    }, job)
    records = append(records, record)
  }

  return records, nil
}

func getJenkins(config JenkinsConfig) (*gojenkins.Jenkins) {
  return gojenkins.CreateJenkins(
    config.JenkinsHost,
    config.JenkinsAPIUser,
    config.JenkinsAPIKey,
  )
}

// gets job information
func getAllJobStats(log *logrus.Logger, jenkins *gojenkins.Jenkins) (jobRecords []map[string]interface{}, err error) {
  jobs, err := jenkins.GetAllJobs()
  if err != nil {
    log.WithError(err).Error("Error getting job statistics")
    return nil, err
  }

  for _, job := range jobs {
    childJobs, childJobsErr := findChildJobs(jenkins, job)
    if childJobsErr == nil {
      jobs = append(jobs, childJobs...)
    }
  }

  for _, job := range jobs {
    jobRecords = append(jobRecords, getJobStats(*job))
  }

  return jobRecords, nil
}

// recursively finds all child jobs for a job
func findChildJobs(jenkins *gojenkins.Jenkins, job *gojenkins.Job) (childJobs []*gojenkins.Job, _ error) {
  if len(job.GetInnerJobsMetadata()) > 0 {
    innerJobs, innerJobsErr := job.GetInnerJobs()
    if innerJobsErr != nil {
      return nil, innerJobsErr
    }

    childJobs = append(childJobs, innerJobs...)

    // find child jobs for each child
    for _, child := range innerJobs {
      childInnerJobs, childInnerJobsErr := findChildJobs(jenkins, child)
      if childInnerJobsErr == nil {
        childJobs = append(childJobs, childInnerJobs...)
      }
    }
  }

  return childJobs, nil
}

// gets stats from an individual job
func getJobStats(job gojenkins.Job) (record map[string]interface{}) {
  record = map[string]interface{}{
    "entity_name": getFullJobName(job),
  }

  healthReport, health := job.Raw.HealthReport, 0
  if healthReport != nil && len(healthReport) > 0 {
    for _, report := range healthReport {
      health += int(report.Score)
    }
    health /= len(healthReport)

    record = mergeMaps(record, map[string]interface{}{
      "jenkins.job.health": health,
    })
  }

  build, buildErr := job.GetLastBuild()
  if buildErr == nil {
    record = mergeMaps(record, map[string]interface{}{
      "jenkins.job.buildNumber": build.GetBuildNumber(),
      "jenkins.job.buildRevision": build.GetRevision(),
      "jenkins.job.buildDate": build.GetTimestamp().Unix(),
      "jenkins.job.buildResult": strings.ToLower(build.GetResult()),
      "jenkins.job.buildDurationSecond": build.GetDuration(),
      "jenkins.job.buildArtifacts": len(build.GetArtifacts()),
    })

    tests, testsErr := build.GetResultSet()
    if testsErr == nil {
      var totalTests int
      for _, suite := range tests.Suites {
        totalTests += len(suite.Cases)
      }
      record = mergeMaps(record, map[string]interface{}{
        "jenkins.job.testsDurationSecond": tests.Duration,
        "jenkins.job.tests": totalTests,
        "jenkins.job.testsSuites": len(tests.Suites),
        "jenkins.job.testsPassed": tests.PassCount,
        "jenkins.job.testsFailed": tests.FailCount,
        "jenkins.job.testsSkipped": tests.SkipCount,
      })
    }
  }

  return record
}

// gets node information
func getAllNodeStats(log *logrus.Logger, jenkins *gojenkins.Jenkins) (nodeRecords []map[string]interface{}, err error) {
  nodes, err := jenkins.GetAllNodes()
  if err != nil {
    log.WithError(err).Error("Error getting node statistics")
    return nil, err
  }

  for _, node := range nodes {
    nodeRecords = append(nodeRecords, getNodeStats(*node))
  }

  return nodeRecords, nil
}

// gets stats from a node
func getNodeStats(node gojenkins.Node) map[string]interface{} {
  return map[string]interface{}{
    "entity_name": node.GetName(),
    "jenkins.node.online": !node.Raw.Offline,
    "jenkins.node.idle": node.Raw.Idle,
    "jenkins.node.executors": len(node.Raw.Executors),
  }
}

// gets the whole name of a job, the hard way
// it is included as "fullName" in the API, but gojenkins Job struct doesn't look for it
func getFullJobName(job gojenkins.Job) string {
  return strings.Trim(strings.Replace(job.Base, "/job/", "/", -1), "/")
}

// merge two maps
func mergeMaps(global map[string]interface{}, specific map[string]interface{}) map[string]interface{} {
  for key, value := range specific {
    global[key] = value
  }
  return global
}
