/*
the datastore package retrieves data from the stackdriver API as well as query of '__Stat_Kind__' from datastore directly.
datastore.go contains the logic to send data to new relic and the datastore query metric/connection. stackdrive.go contains
the stackdriver API metrics/connection.
*/

package datastore

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/GannettDigital/go-newrelic-plugin/helpers"
	"github.com/Sirupsen/logrus"
	"github.com/buger/jsonparser"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/monitoring/v3"
	"google.golang.org/api/option"
)

const (
	// NAME - name of plugin
	NAME string = "datastore"

	// ProtocolVersion -
	ProtocolVersion string = "1"
)

var base64Config string

var stackdriverEndpoints = []string{
	"datastore.googleapis.com/api/request_count",
	"datastore.googleapis.com/index/write_count",
	//"datastore.googleapis.com/entity/read_sizes", --TODO add distribution data
	//"datastore.googleapis.com/entity/write_sizes",
}

//DatastoreKind represents the fields for a datastore Query
type DatastoreKind struct {
	BuiltinIndexBytes   int       `datastore:"builtin_index_bytes"`
	BuiltinIndexCount   int       `datastore:"builtin_index_count"`
	CompositeIndexBytes int       `datastore:"composite_index_bytes"`
	CompositeIndexCount int       `datastore:"composite_index_count"`
	EntityBytes         int       `datastore:"entity_bytes"`
	Bytes               int       `datastore:"bytes"`
	Count               int       `datastore:"count"`
	KindName            string    `datastore:"kind_name"`
	Timestamp           time.Time `datastore:"timestamp"`
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

//stores datastoreclient
type Client struct {
	Dsc DatastoreClient
}

//DatastoreClient is used for testing purposes
type DatastoreClient interface {
	GetAll(ctx context.Context, q *datastore.Query, dst interface{}) (keys []*datastore.Key, err error)
}

// InventoryData is the data type for inventory data produced by a plugin data
// source and emitted to the agent's inventory data store
type InventoryData map[string]interface{}

// MetricData is the data type for events produced by a plugin data source and
// emitted to the agent's metrics data store
type MetricData map[string]interface{}

// EventData is the data type for single shot events
type EventData map[string]interface{}

//StackdriverMetric represents fields for stackdriver returns
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

	c, projectId, err := NewClient()
	if err != nil {
		log.Fatal(err)
	}

	//add query metrics
	kinds, err := DatastoreStatKindQueryResult(c.Dsc)
	if err != nil {
		log.Fatal(err)
	}
	result := DatastoreData(kinds, projectId)

	for _, metricResult := range result {
		data.Metrics = append(data.Metrics, metricResult)
	}

	//add stackdriver metrics
	for _, metric := range stackdriverEndpoints {
		resp, err := StackdriverResp(projectId, metric)
		if err != nil {
			log.Fatal(err)
		}

		result, err := StackdriverData(resp)
		if err != nil {
			log.Fatal(err)
		}
		for _, metricResult := range result {
			data.Metrics = append(data.Metrics, metricResult)
		}
	}

	fatalIfErr(log, helpers.OutputJSON(data, prettyPrint))
}

//NewClient() creates a client for datastore, it primarily exists for testing purposes
func NewClient() (Client, string, error) {
	dsClient, projectId, err := ConnectDatastore(base64Config)
	if err != nil {
		return Client{}, "", err
	}

	return Client{
		dsClient,
	}, projectId, nil
}

//DatastoreStatKindQueryResult performs a Query against the datastore and return values using __Stat_Kind__
func DatastoreStatKindQueryResult(ds DatastoreClient) ([]DatastoreKind, error) {
	q := datastore.NewQuery("__Stat_Kind__").Order("kind_name")

	var kinds []DatastoreKind

	ctx := context.Background()
	_, err := ds.GetAll(ctx, q, &kinds)

	if err != nil {
		return nil, err
	}

	return kinds, nil
}

//DatastoreData converts a []DatastoreKind to a []map[string]interface{} to be used for the final output to new relic
func DatastoreData(kinds []DatastoreKind, projectID string) []map[string]interface{} {
	var datastoreData []map[string]interface{}
	for _, k := range kinds {
		datastoreData = append(datastoreData, map[string]interface{}{
			"event_type":                         "DatastoreSample",
			"provider":                           "datastoreQuery",
			"datastoreQuery.builtinIndexBytes":   k.BuiltinIndexBytes,
			"datastoreQuery.builtinIndexCount":   k.BuiltinIndexCount,
			"datastoreQuery.compositeIndexBytes": k.CompositeIndexBytes,
			"datastoreQuery.compositeIndexCount": k.CompositeIndexCount,
			"datastoreQuery.entityBytes":         k.EntityBytes,
			"datastoreQuery.bytes":               k.Bytes,
			"datastoreQuery.count":               k.Count,
			"datastoreQuery.kindName":            k.KindName,
			"datastoreQuery.projectId":           projectID,
			"datastoreQuery.timestamp":           k.Timestamp.Unix(),
		})
	}
	return datastoreData
}

func fatalIfErr(log *logrus.Logger, err error) {
	if err != nil {
		log.WithError(err).Fatal("can't continue")
	}
}

// ConnectDatastore establishes a datastore.Client from a base64 encoding JSON credentials file.
func ConnectDatastore(base64Config string) (*datastore.Client, string, error) {
	jsonConfig, err := base64.StdEncoding.DecodeString(base64Config)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode datastore credentials: %v", err)
	}
	projectId, err := jsonparser.GetString(jsonConfig, "project_id")
	if err != nil {
		return nil, "", fmt.Errorf("failed to determine project_id from credentials file: %v", err)
	}

	ctx := context.Background()
	creds, err := google.CredentialsFromJSON(ctx, jsonConfig, datastore.ScopeDatastore)
	if err != nil {
		return nil, "", err
	}
	c, err := datastore.NewClient(ctx, projectId, option.WithCredentials(creds))
	if err != nil {
		return nil, "", err
	}
	return c, projectId, err
}

//StackdriverResp gets the data of the wanted metric from the stackdriver API. Start time is set at -3 minutes to act as a Timestamp
//as data from stackdriver is always 3 minutes old and refreshed every 1 minute. This timing also ensures we only ever get 1 point
//back at a time.
func StackdriverResp(projectId string, metric string) (*monitoring.ListTimeSeriesResponse, error) {
	ctx := context.Background()
	jsonConfig, err := base64.StdEncoding.DecodeString(base64Config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode datastore credentials: %v", err)
	}

	projectId, err = jsonparser.GetString(jsonConfig, "project_id")
	if err != nil {
		return nil, fmt.Errorf("failed to determine project_id from credentials file: %v", err)
	}

	creds, err := google.CredentialsFromJSON(ctx, jsonConfig, monitoring.MonitoringScope)

	hc := oauth2.NewClient(ctx, creds.TokenSource)

	s, err := monitoring.New(hc)
	if err != nil {
		return nil, err
	}

	startTime := time.Now().UTC().Add(time.Minute * -3)
	endTime := time.Now().UTC()

	resp, err := s.Projects.TimeSeries.List(projectResource(projectId)).
		Filter(fmt.Sprintf("metric.type=\"%s\"", metric)).
		IntervalStartTime(startTime.Format(time.RFC3339)).
		IntervalEndTime(endTime.Format(time.RFC3339)).
		Do()

	if err != nil {
		return nil, err
	}

	return resp, nil

}

//StackdriverData converts a ListTimeSeriesResponse to a []map[string]interface{} to be used for the final output to be sent to new relic
func StackdriverData(resp *monitoring.ListTimeSeriesResponse) ([]map[string]interface{}, error) {
	var stackdriverMetricBody StackdriverMetric
	var metricResult []map[string]interface{}

	b, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &stackdriverMetricBody)

	if err != nil {
		return nil, err
	}

	for _, item := range stackdriverMetricBody.TimeSeries {
		metricResult = append(metricResult, map[string]interface{}{
			"event_type":                        "DatastoreSample",
			"provider":                          "datastoreStackdriver",
			"datastoreStackdriver.apiMethod":    item.Metric.Labels.ApiMethod,
			"datastoreStackdriver.responseCode": item.Metric.Labels.ResponseCode,
			"datastoreStackdriver.metricType":   item.Metric.Type,
			"datastoreStackdriver.metricKind":   item.MetricKind,
			"datastoreStackdriver.timestamp":    item.Points[0].Interval.StartTime.Unix(),
			"datastoreStackdriver.value":        item.Points[0].Value.Int64Value,
			"datastoreStackdriver.projectId":    item.Resource.Labels.ProjectID,
			"datastoreStackdriver.resourceType": item.Resource.Type,
		})
	}

	return metricResult, nil
}

func projectResource(projectID string) string {
	return "projects/" + projectID
}

