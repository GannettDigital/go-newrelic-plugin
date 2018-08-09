package datastore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/GannettDigital/go-newrelic-plugin/helpers"
	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/monitoring/v3"
)

var stackdriverEndpoints = []string{
	"datastore.googleapis.com/api/request_count",
	"datastore.googleapis.com/index/write_count",
	//"datastore.googleapis.com/entity/read_sizes", --TODO add distribution data
	//"datastore.googleapis.com/entity/write_sizes",
}

//fields for datastore Query
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

//fields for stackdriver returns
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
	Dsc DatastoreImpl
	Sdc StackDriverImpl
}


type DatastoreImpl interface {
	GetAll(ctx context.Context, q *datastore.Query, dst interface{}) (keys []*datastore.Key, err error)
}

type StackDriverImpl interface{
	List(name string) *monitoring.ProjectsTimeSeriesListCall
	// add methods that are used by stackDriver. You likely need multiple interfaces sense it does chaining like List.Filter....DO()
}



func NewClient(projectId string) Client {

	ctx := context.Background()
	dsClient, err := datastore.NewClient(ctx, projectId)
	if err != nil {
		panic("unable to set client")
	}

	hc, err := google.DefaultClient(ctx, monitoring.MonitoringScope)
	if err != nil {
		panic("unable to set client")
	}
	s, err := monitoring.New(hc)
	if err != nil {
		panic("unable to set client")
	}

	return Client{
		dsClient,
		s.Projects.TimeSeries,
	}
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

// NAME - name of plugin
const NAME string = "datastore"

// PROVIDER -
const PROVIDER string = "datastore" //we might want to make this an env tied to nginx version or app name maybe...

// ProtocolVersion -
const ProtocolVersion string = "1"


// I would not test this method
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

	//keyFile,err :=ioutil.ReadFile("/var/secrets/google/key.json")
	keyFile, err := ioutil.ReadFile("/Users/jstorer/Downloads/gannett-api-services-stage-e.json")
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(keyFile, &keyData)
	if err != nil {
		log.Fatal(err)
	}

	//os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/var/secrets/google/key.json")
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/Users/jstorer/Downloads/gannett-api-services-stage-e.json")

	projectID := keyData.ProjectID

	c := NewClient(projectID)


	//add stackdriver metrics
	for _, metric := range stackdriverEndpoints {
		resp, err := getStackdriverResp(c.Sdc, projectID, metric)
		if err != nil {
			log.Fatal(err)
		}

		result, err := getStackdriverData(resp)
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

// I would not test this method
func getStackdriverResp(sdl StackDriverImpl , projectID string, metric string) (*monitoring.ListTimeSeriesResponse, error) {
	startTime := time.Now().UTC().Add(time.Minute * -3)
	endTime := time.Now().UTC()

	resp, err := sdl.List(projectResource(projectID)).
		Filter(fmt.Sprintf("metric.type=\"%s\"", metric)).
		IntervalStartTime(startTime.Format(time.RFC3339)).
		IntervalEndTime(endTime.Format(time.RFC3339)).
		Do()

	if err != nil {
		return nil, err
	}

	return resp, nil

}


// To test this, I would pass in a filled out *monitoring.ListTimeSeriesResponse with the data you know
// then, check the results of this data validating we are setting the correct proprties.. ie datastoreStackdriver.resourceType to the value you would expect
func getStackdriverData(resp *monitoring.ListTimeSeriesResponse) ([]map[string]interface{}, error) {
	var stackdriverMetricBody StackdriverMetric
	var metricResult []map[string]interface{}

	err := json.Unmarshal(formatResource(resp), &stackdriverMetricBody)
	if err != nil {
		return nil, err
	}

	for id := range stackdriverMetricBody.TimeSeries {
		metricResult = append(metricResult, map[string]interface{}{
			"event_type":                        "DatastoreSample",
			"provider":                          "datastoreStackdriver",
			"datastoreStackdriver.apiMethod":    stackdriverMetricBody.TimeSeries[id].Metric.Labels.ApiMethod,
			"datastoreStackdriver.responseCode": stackdriverMetricBody.TimeSeries[id].Metric.Labels.ResponseCode,
			"datastoreStackdriver.metricType":   stackdriverMetricBody.TimeSeries[id].Metric.Type,
			"datastoreStackdriver.metricKind":   stackdriverMetricBody.TimeSeries[id].MetricKind,
			"datastoreStackdriver.timestamp":    stackdriverMetricBody.TimeSeries[id].Points[0].Interval.StartTime.Unix(),
			"datastoreStackdriver.value":        stackdriverMetricBody.TimeSeries[id].Points[0].Value.Int64Value,
			"datastoreStackdriver.projectId":    stackdriverMetricBody.TimeSeries[id].Resource.Labels.ProjectID,
			"datastoreStackdriver.resourceType": stackdriverMetricBody.TimeSeries[id].Resource.Type,
		})
	}

	return metricResult, nil
}

func getDatastoreQueryResult(ds DatastoreImpl) ([]*DatastoreKind, error) {

	q := datastore.NewQuery("__Stat_Kind__").Order("kind_name")

	kinds := []*DatastoreKind{}

	ctx := context.Background()
	_, err := ds.GetAll(ctx, q, &kinds)

	if err != nil {
		return nil, err
	}

	return kinds, nil
}

func getDatastoreData(kinds []*DatastoreKind, projectID string) []map[string]interface{} {
	var queryResult []map[string]interface{}
	for _, k := range kinds {
		queryResult = append(queryResult, map[string]interface{}{
			"event_type":               "DatastoreSample",
			"provider":                 "datastoreQuery",
			"datastoreQuery.timestamp": time.Now().UTC().Add(time.Minute * -3).Unix(),
			"datastoreQuery.kindName":  k.KindName,
			"datastoreQuery.count":     k.Count,
			"datastoreQuery.bytes":     k.Bytes,
			"datastoreQuery.projectId": projectID,
		})
	}
	return queryResult
}

func createService(ctx context.Context) (*monitoring.Service, error) {
	hc, err := google.DefaultClient(ctx, monitoring.MonitoringScope)
	if err != nil {
		return nil, err
	}
	s, err := monitoring.New(hc)
	if err != nil {
		return nil, err
	}
	return s, nil
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
