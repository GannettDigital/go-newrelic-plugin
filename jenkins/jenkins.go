package jenkins

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bndr/gojenkins"
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

type JobMetric struct {
	EntityName          string    `json:"entity_name"`
	EventType           string    `json:"event_type"`
	Provider            string    `json:"provider"`
	Health              int       `json:"jenkins.job.health"`
	BuildNumber         int       `json:"jenkins.job.buildNumber"`
	BuildRevision       string    `json:"jenkins.job.buildRevision"`
	BuildDate           time.Time `json:"jenkins.job.buildDate"`
	BuildResult         string    `json:"jenkins.job.buildResult"`
	BuildDurationSecond int       `json:"jenkins.job.buildDurationSecond"`
	BuildArtifacts      int       `json:"jenkins.job.buildArtifacts"`
	TestsDurationSecond int       `json:"jenkins.job.testsDurationSecond"`
	TestsSuites         int       `json:"jenkins.job.testsSuites"`
	Tests               int       `json:"jenkins.job.tests"`
	TestsPassed         int       `json:"jenkins.job.testsPassed"`
	TestsFailed         int       `json:"jenkins.job.testsFailed"`
	TestsSkipped        int       `json:"jenkins.job.testsSkipped"`
}

type NodeMetric struct {
	EntityName string `json:"entity_name"`
	EventType  string `json:"event_type"`
	Provider   string `json:"provider"`
	Online     bool   `json:"jenkins.node.online"`
	Idle       bool   `json:"jenkins.node.idle"`
	Executors  int    `json:"jenkins.node.executors"`
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
		JenkinsHost:    os.Getenv("JENKINS_HOST"),
		JenkinsAPIUser: os.Getenv("JENKINS_API_USER"),
		JenkinsAPIKey:  os.Getenv("JENKINS_API_KEY"),
	}
	validErr := validateConfig(log, config)
	if validErr != nil {
		log.Fatalf("config: %v\n", validErr)
	}

	jenkins, jenkinsErr := getJenkins(config).Init()
	if jenkinsErr != nil {
		log.WithError(jenkinsErr).Fatal("Error connecting to Jenkins")
	}

	metrics, metricsErr := getMetrics(log, jenkins)
	fatalIfErr(log, metricsErr)

	data.Metrics = append(data.Metrics, metrics...)
	fatalIfErr(log, OutputJSON(data, prettyPrint))
}

func validateConfig(log *logrus.Logger, config JenkinsConfig) error {
	if config.JenkinsHost == "" {
		return errors.New("JENKINS_HOST must be set")
	}
	if config.JenkinsAPIUser != "" && config.JenkinsAPIKey == "" {
		return errors.New("you must also set JENKINS_API_KEY when JENKINS_API_USER is set")
	}
	if config.JenkinsAPIUser == "" && config.JenkinsAPIKey != "" {
		return errors.New("you must also set JENKINS_API_USER when JENKINS_API_KEY is set")
	}
	return nil
}

func fatalIfErr(log *logrus.Logger, err error) {
	if err != nil {
		log.WithError(err).Fatal("can't continue")
	}
}

func getMetrics(log *logrus.Logger, jenkins *gojenkins.Jenkins) ([]MetricData, error) {
	var records []MetricData

	jobData, jobDataErr := getAllJobStats(log, jenkins)
	if jobDataErr != nil {
		return nil, jobDataErr
	}
	for _, job := range jobData {
		records = append(records, MetricData{
			"entity_name":                     job.EntityName,
			"event_type":                      "DatastoreSample",
			"provider":                        "jenkins.job",
			"jenkins.job.health":              job.Health,
			"jenkins.job.buildNumber":         job.BuildNumber,
			"jenkins.job.buildRevision":       job.BuildRevision,
			"jenkins.job.buildDate":           job.BuildDate,
			"jenkins.job.buildResult":         job.BuildResult,
			"jenkins.job.buildDurationSecond": job.BuildDurationSecond,
			"jenkins.job.buildArtifacts":      job.BuildArtifacts,
			"jenkins.job.testsDurationSecond": job.TestsDurationSecond,
			"jenkins.job.testsSuites":         job.TestsSuites,
			"jenkins.job.tests":               job.Tests,
			"jenkins.job.testsPassed":         job.TestsPassed,
			"jenkins.job.testsFailed":         job.TestsFailed,
			"jenkins.job.testsSkipped":        job.TestsSkipped,
		})
	}

	nodeData, nodeDataErr := getAllNodeStats(log, jenkins)
	if nodeDataErr != nil {
		return nil, nodeDataErr
	}
	for _, node := range nodeData {
		records = append(records, MetricData{
			"entity_name":            node.EntityName,
			"event_type":             "LoadBalancerSample",
			"provider":               "jenkins.node",
			"jenkins.node.online":    node.Online,
			"jenkins.node.idle":      node.Idle,
			"jenkins.node.executors": node.Executors,
		})
	}

	return records, nil
}

func getJenkins(config JenkinsConfig) *gojenkins.Jenkins {
	return gojenkins.CreateJenkins(
		config.JenkinsHost,
		config.JenkinsAPIUser,
		config.JenkinsAPIKey,
	)
}

// gets job information
func getAllJobStats(log *logrus.Logger, jenkins *gojenkins.Jenkins) ([]JobMetric, error) {
	var jobRecords []JobMetric

	jobs, jobsErr := jenkins.GetAllJobs()
	if jobsErr != nil {
		log.WithError(jobsErr).Error("Error getting job statistics")
		return jobRecords, jobsErr
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
func findChildJobs(jenkins *gojenkins.Jenkins, job *gojenkins.Job) ([]*gojenkins.Job, error) {
	var childJobs []*gojenkins.Job

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
func getJobStats(job gojenkins.Job) JobMetric {
	record := JobMetric{
		EntityName: getFullJobName(job),
	}

	health := 0
	healthReport := job.Raw.HealthReport
	if healthReport != nil && len(healthReport) > 0 {
		for _, report := range healthReport {
			health += int(report.Score)
		}
		health /= len(healthReport)
		record.Health = health
	}

	build, buildErr := job.GetLastBuild()
	if buildErr == nil {
		record.BuildNumber = int(build.GetBuildNumber())
		record.BuildRevision = build.GetRevision()
		record.BuildDate = build.GetTimestamp()
		record.BuildResult = strings.ToLower(build.GetResult())
		record.BuildDurationSecond = int(build.GetDuration())
		record.BuildArtifacts = len(build.GetArtifacts())

		tests, testsErr := build.GetResultSet()
		if testsErr == nil {
			totalTests := 0
			for _, suite := range tests.Suites {
				totalTests += len(suite.Cases)
			}
			record.TestsDurationSecond = int(tests.Duration)
			record.Tests = totalTests
			record.TestsSuites = len(tests.Suites)
			record.TestsPassed = int(tests.PassCount)
			record.TestsFailed = int(tests.FailCount)
			record.TestsSkipped = int(tests.SkipCount)
		}
	}

	return record
}

// gets node information
func getAllNodeStats(log *logrus.Logger, jenkins *gojenkins.Jenkins) ([]NodeMetric, error) {
	var nodeRecords []NodeMetric

	nodes, nodesErr := jenkins.GetAllNodes()
	if nodesErr != nil {
		log.WithError(nodesErr).Error("Error getting node statistics")
		return nil, nodesErr
	}

	for _, node := range nodes {
		nodeRecords = append(nodeRecords, getNodeStats(*node))
	}

	return nodeRecords, nil
}

// gets stats from a node
func getNodeStats(node gojenkins.Node) NodeMetric {
	return NodeMetric{
		EntityName: node.GetName(),
		Online:     !node.Raw.Offline,
		Idle:       node.Raw.Idle,
		Executors:  len(node.Raw.Executors),
	}
}

// gets the whole name of a job, the hard way
// it is included as "fullName" in the API, but gojenkins Job struct doesn't look for it
func getFullJobName(job gojenkins.Job) string {
	return strings.Trim(strings.Replace(job.Base, "/job/", "/", -1), "/")
}
