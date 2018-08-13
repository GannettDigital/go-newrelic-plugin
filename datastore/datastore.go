/*
the datastore package retrieves data from the stackdriver API as well as query of '__Stat_Kind__' from datastore directly.
*/

package datastore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/buger/jsonparser"
	"github.com/GannettDigital/go-newrelic-plugin/helpers"
	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/monitoring/v3"
	"encoding/base64"
	"google.golang.org/api/option"
)

const (
	// NAME - name of plugin
	NAME string = "datastore"
	// ProtocolVersion -
	ProtocolVersion string = "1"
)

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

type Client struct {
	Dsc DatastoreInterface
}

//DatastoreInterface is used for testing purposes
type DatastoreInterface interface {
	GetAll(ctx context.Context, q *datastore.Query, dst interface{}) (keys []*datastore.Key, err error)
}

type KeyData struct {
	AuthProviderX509CertURL string  `json:"auth_provider_x509_cert_url"`
	AuthURI                 string  `json:"auth_uri"`
	ClientEmail             string  `json:"client_email"`
	ClientID                float64 `json:"client_id,string"`
	ClientX509CertURL       string  `json:"client_x509_cert_url"`
	PrivateKey              string  `json:"private_key"`
	PrivateKeyID            string  `json:"private_key_id"`
	ProjectID               string  `json:"project_id"`
	TokenURI                string  `json:"token_uri"`
	Type                    string  `json:"type"`
}

// InventoryData is the data type for inventory data produced by a plugin data
// source and emitted to the agent's inventory data store
type InventoryData map[string]interface{}

// MetricData is the data type for events produced by a plugin data source and
// emitted to the agent's metrics data store
type MetricData map[string]interface{}

// EventData is the data type for single shot events
type EventData map[string]interface{}

func Run(log *logrus.Logger, prettyPrint bool, version string) {
	var data = PluginData{
		Name:            NAME,
		ProtocolVersion: ProtocolVersion,
		PluginVersion:   version,
		Inventory:       make(map[string]InventoryData),
		Metrics:         make([]MetricData, 0),
		Events:          make([]EventData, 0),
	}

	var keyData KeyData

	keyFile, err := ioutil.ReadFile("/var/secrets/google/key.json")

	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(keyFile, &keyData)
	if err != nil {
		log.Fatal(err)
	}

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/var/secrets/google/key.json")

	projectID := keyData.ProjectID
	c,err := NewClient(projectID)
	if err != nil {
		log.Fatal(err)
	}

	//add stackdriver metrics
	for _, metric := range stackdriverEndpoints {
		resp, err := stackdriverResp(projectID, metric)
		if err != nil {
			log.Fatal(err)
		}

		result, err := stackdriverData(resp)
		if err != nil {
			log.Fatal(err)
		}
		for _, metricResult := range result {
			data.Metrics = append(data.Metrics, metricResult)
		}
	}

	//add query metrics
	kinds, err := getDatastoreQueryResult(c.Dsc)
	if err != nil {
		log.Fatal(err)
	}
	result := getDatastoreData(kinds, projectID)

	for _, metricResult := range result {
		data.Metrics = append(data.Metrics, metricResult)
	}

	fatalIfErr(log, helpers.OutputJSON(data, prettyPrint))
}

func stackdriverResp(projectID string, metric string) (*monitoring.ListTimeSeriesResponse, error) {
	ctx := context.Background()

	hc,err := google.DefaultClient(ctx,monitoring.MonitoringScope)
	if err != nil {
		return nil, err
	}

	s, err := monitoring.New(hc)
	if err != nil {
		return nil, err
	}

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

func stackdriverData(resp *monitoring.ListTimeSeriesResponse) ([]map[string]interface{}, error) {
	var stackdriverMetricBody StackdriverMetric
	var metricResult []map[string]interface{}

	err := json.Unmarshal(formatResource(resp), &stackdriverMetricBody)
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

func getDatastoreQueryResult(ds DatastoreInterface) ([]DatastoreKind, error) {
	q := datastore.NewQuery("__Stat_Kind__").Order("kind_name")

	var kinds []DatastoreKind

	ctx := context.Background()
	_, err := ds.GetAll(ctx, q, &kinds)

	if err != nil {
		return nil, err
	}

	return kinds, nil
}

func getDatastoreData(kinds []DatastoreKind, projectID string) []map[string]interface{} {
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

func NewClient(projectId string) (Client,error) {

	ctx := context.Background()
	dsClient, err := datastore.NewClient(ctx, projectId)
	if err != nil {
		return Client{},err
	}

	hc, err := google.DefaultClient(ctx, monitoring.MonitoringScope)
	if err != nil {
		return Client{},err
	}
	s, err := monitoring.New(hc)
	if err != nil {
		return Client{},err
	}

	return Client{
		dsClient
	},nil
}

func fatalIfErr(log *logrus.Logger, err error) {
	if err != nil {
		log.WithError(err).Fatal("can't continue")
	}
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

// ConnectDatastore establishes a datastore.Client from a base64 encoding JSON credentials file.
func ConnectDatastore(base64Config string) (*datastore.Client, error) {
	jsonConfig, err := base64.StdEncoding.DecodeString(base64Config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode datastore credentials: %v", err)
	}
	projectID, err := jsonparser.GetString(jsonConfig, "project_id")
	if err != nil {
		return nil, fmt.Errorf("failed to determine project_id from credentials file: %v", err)
	}

	ctx := context.Background()
	creds, err := google.CredentialsFromJSON(ctx, jsonConfig, datastore.ScopeDatastore)
	if err != nil {
		return nil, err
	}

	return datastore.NewClient(ctx, projectID, option.WithCredentials(creds))
}
