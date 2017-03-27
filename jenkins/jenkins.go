package jenkins

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/GannettDigital/go-newrelic-plugin/helpers"
	"github.com/Sirupsen/logrus"
	"github.com/bndr/gojenkins"
)

// CollectorName - the name of this thing
const CollectorName string = "jenkins"

// ProviderName - what app is sending the data
const ProviderName string = "jenkins"

// ProtocolVersion - nr-infra protocol version
const ProtocolVersion string = "1"

// Config stores the config to connect to the Jenkins master from which data will be retrieved
type Config struct {
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

// JobMetric stores metrics from jobs
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

// NodeMetric stores metrics from build nodes
type NodeMetric struct {
	EntityName string `json:"entity_name"`
	EventType  string `json:"event_type"`
	Provider   string `json:"provider"`
	Online     bool   `json:"jenkins.node.online"`
	Idle       bool   `json:"jenkins.node.idle"`
	Executors  int    `json:"jenkins.node.executors"`
}

// Run connects to Jenkins, grabs data, and prints it to stdout
func Run(log *logrus.Logger, prettyPrint bool, version string) {

	// Initialize the output structure
	var data = PluginData{
		Name:            CollectorName,
		ProtocolVersion: ProtocolVersion,
		PluginVersion:   version,
		Inventory:       make(map[string]InventoryData),
		Metrics:         make([]MetricData, 0),
		Events:          make([]EventData, 0),
		Status:          "ok",
	}

	// get config from env vars
	var config = Config{
		JenkinsHost:    os.Getenv("JENKINS_HOST"),
		JenkinsAPIUser: os.Getenv("JENKINS_API_USER"),
		JenkinsAPIKey:  os.Getenv("JENKINS_API_KEY"),
	}
	validErr := validateConfig(config)
	if validErr != nil {
		log.WithError(validErr).Error("Error with configuration")
		return
	}

	jenkins, jenkinsErr := getJenkins(config).Init()
	if jenkinsErr != nil {
		log.WithError(jenkinsErr).Error("Error connecting to Jenkins")
		return
	}

	metrics, metricsErr := getMetrics(log, jenkins)
	if metricsErr != nil {
		log.WithError(metricsErr).Error("Error collecting metrics")
		return
	}
	data.Metrics = append(data.Metrics, metrics...)

	outputErr := helpers.OutputJSON(data, prettyPrint)
	if outputErr != nil {
		log.WithError(outputErr).Error("Error formatting output JSON")
		return
	}
}

func validateConfig(config Config) error {
	if config.JenkinsHost == "" {
		return fmt.Errorf("JENKINS_HOST must be set")
	}
	if config.JenkinsAPIUser != "" && config.JenkinsAPIKey == "" {
		return fmt.Errorf("You must also set JENKINS_API_KEY if JENKINS_API_USER is set")
	}
	if config.JenkinsAPIUser == "" && config.JenkinsAPIKey != "" {
		return fmt.Errorf("You must also set JENKINS_API_USER if JENKINS_API_KEY is set")
	}
	return nil
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

func getJenkins(config Config) *gojenkins.Jenkins {
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
