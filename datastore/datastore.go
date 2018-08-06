package datastore

import (
	"encoding/json"
	"fmt"
	"time"
	"strconv"

	"cloud.google.com/go/datastore"
	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/monitoring/v3"
	"github.com/GannettDigital/go-newrelic-plugin/helpers"
)

var stackdriverEndpoints = []string{
	"datastore.googleapis.com/api/request_count",
	"datastore.googleapis.com/index/write_count",
	//"datastore.googleapis.com/entity/read_sizes",
	//"datastore.googleapis.com/entity/write_sizes", -- TODO create appropriate sized and number of buckets
}

type DatastoreKind struct {
	KindName            string    `datastore:"kind_name"`
	EntityBytes         int       `datastore:"entity_bytes"`
	BuiltinIndexBytes   int       `datastore:"builtin_index_bytes"`
	BuiltinIndexCount   int       `datastore:"builtin_index_count"`
	CompositeIndexBytes int       `datastore:"composite_index_bytes"`
	CompositeIndexCount int       `datastore:"composite_index_count"`
	Timestamp           time.Time `datastore:"timestamp"`
	Count               int       `datastore:"count"`
	Bytes               int       `datastore:"bytes"`
}

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

// InventoryData is the data type for inventory data produced by a plugin data
// source and emitted to the agent's inventory data store
type InventoryData map[string]interface{}

// MetricData is the data type for events produced by a plugin data source and
// emitted to the agent's metrics data store
type MetricData map[string]interface{}

// EventData is the data type for single shot events
type EventData map[string]interface{}

// NAME - name of plugin
const NAME string = "datastore"

// PROVIDER -
const PROVIDER string = "datastore" //we might want to make this an env tied to nginx version or app name maybe...

// ProtocolVersion -
const ProtocolVersion string = "1"

type StackdriverMetric struct {
	TimeSeries []struct {
		Metric struct {
			Labels struct {
				ApiMethod    string `json:"api_method"`
				ResponseCode string `json:"response_code"`
			} `json:"labels"`
			Type string `json:"type"`
		} `json:"metric"`
		MetricKind string `json:"metricKind"`
		Points     []struct {
			Interval struct {
				EndTime   time.Time `json:"endTime"`
				StartTime time.Time `json:"startTime"`
			} `json:"interval"`
			Value struct {
				Int64Value int64 `json:"int64Value,string"`
			} `json:"value"`
		} `json:"points"`
		Resource struct {
			Labels struct {
				ModuleID  string `json:"module_id"`
				ProjectID string `json:"project_id"`
				VersionID string `json:"version_id"`
			} `json:"labels"`
			Type string `json:"type"`
		} `json:"resource"`
		ValueType string `json:"valueType"`
	} `json:"timeSeries"`
}

func Run(log *logrus.Logger, prettyPrint bool, version string) {
	var data = PluginData{
		Name:            NAME,
		ProtocolVersion: ProtocolVersion,
		PluginVersion:   version,
		Inventory:       make(map[string]InventoryData),
		Metrics:         make([]MetricData, 0),
		Events:          make([]EventData, 0),
	}

	ctx := context.Background()
	projectID := "gannett-api-services-stage"

	//stackdriver metrics
	hc, err := google.DefaultClient(ctx, monitoring.MonitoringScope)

	if err != nil {
		log.Fatal(err)
	}

	s, err := monitoring.New(hc)

	if err != nil {
		log.Fatal(err)
	}

	//loop through datastore stackdriverEndpoints
	for _, metric := range stackdriverEndpoints {

		resp, err := listTimeSeries(s, projectID, metric)

		if err != nil {
			log.Fatal(err)
		}

		var requestCountBody StackdriverMetric

		json.Unmarshal(formatResource(resp), &requestCountBody)

		//fmt.Printf("%s",formatResource(resp))
		for id := range requestCountBody.TimeSeries {
			data.Metrics = append(data.Metrics, map[string]interface{}{
				"event_type":                        "DatastoreSample",
				"provider":                          PROVIDER,
				"datastoreStackdriver.apiMethod":    requestCountBody.TimeSeries[id].Metric.Labels.ApiMethod,
				"datastoreStackdriver.responseCode": requestCountBody.TimeSeries[id].Metric.Labels.ResponseCode,
				"datastoreStackdriver.metricType":   requestCountBody.TimeSeries[id].Metric.Type,
				"datastoreStackdriver.metricKind":   requestCountBody.TimeSeries[id].MetricKind,
				"datastoreStackriver.time":          requestCountBody.TimeSeries[id].Points[0].Interval.EndTime.String(),
				"datastoreStackdriver.value":        strconv.FormatInt(requestCountBody.TimeSeries[0].Points[0].Value.Int64Value, 10),
				"datastoreStackdriver.projectId":    requestCountBody.TimeSeries[id].Resource.Labels.ProjectID,
				"datastoreStackdriver.resourceType": requestCountBody.TimeSeries[id].Resource.Type,
			})
		}
	}

	//datastore statistics
	dsClient, err := datastore.NewClient(ctx, "gannett-api-services-stage")
	if err != nil {
		log.Fatal("Error connecting to datastore")
	}

	q := datastore.NewQuery("__Stat_Kind__").Order("kind_name")

	kinds := []*DatastoreKind{}

	_, err = dsClient.GetAll(ctx, q, &kinds)

	if err != nil {
		log.Fatal(err)
	}

	for _, k := range kinds {
		fmt.Printf("\nkind %q\t%d entries\t%d bytes\n", k.KindName, k.Count, k.Bytes)
		data.Metrics = append(data.Metrics, map[string]interface{}{
			"event_type":              "DatastoreSample",
			"provider":                PROVIDER,
			"datastoreQuery.time":     time.Now().UTC(),
			"datastoreQuery.kindName": k.KindName,
			"datastoreQuery.count":    k.Count,
			"datastoreQuery.bytes":    k.Bytes,
		})
	}

	fatalIfErr(log, helpers.OutputJSON(data,prettyPrint))
}

func fatalIfErr(log *logrus.Logger, err error) {
	if err != nil {
		log.WithError(err).Fatal("can't continue")
	}
}

func listTimeSeries(s *monitoring.Service, projectID string, metric string) (*monitoring.ListTimeSeriesResponse, error) {
	startTime := time.Now().UTC().Add(time.Minute * -3)
	endTime := time.Now().UTC()

	resp, err := s.Projects.TimeSeries.List(projectResource(projectID)).
		Filter(fmt.Sprintf("metric.type=\"%s\"", metric)).
		IntervalStartTime(startTime.Format(time.RFC3339)).
		IntervalEndTime(endTime.Format(time.RFC3339)).
		Do()

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func projectResource(projectID string) string {
	return "projects/" + projectID
}

// formatResource marshals a response objects as JSON.
func formatResource(resource interface{}) []byte {
	b, err := json.MarshalIndent(resource, "", "    ")
	if err != nil {
		panic(err)
	}
	return b
}
