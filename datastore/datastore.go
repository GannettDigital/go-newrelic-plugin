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
	"golang.org/x/oauth2/google"
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

type Client struct {
	Dsc DatastoreInterface
}

//DatastoreInterface is used for testing purposes
type DatastoreInterface interface {
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
	kinds, err := getDatastoreQueryResult(c.Dsc)
	if err != nil {
		log.Fatal(err)
	}
	result := getDatastoreData(kinds,projectId)

	for _, metricResult := range result {
		data.Metrics = append(data.Metrics, metricResult)
	}

	//add stackdriver metrics
	for _, metric := range stackdriverEndpoints {
		resp, err := stackdriverResp(projectId, metric)
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

	fatalIfErr(log, helpers.OutputJSON(data, prettyPrint))
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

func NewClient() (Client,string, error) {
	dsClient,projectId,err:=ConnectDatastore(base64Config)
	if err != nil {
		return Client{},"", err
	}

	return Client{
		dsClient,
	},projectId, nil
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
func ConnectDatastore(base64Config string) (*datastore.Client,string, error) {
	jsonConfig, err := base64.StdEncoding.DecodeString(base64Config)
	if err != nil {
		return nil,"", fmt.Errorf("failed to decode datastore credentials: %v", err)
	}
	projectId, err := jsonparser.GetString(jsonConfig, "project_id")
	if err != nil {
		return nil,"", fmt.Errorf("failed to determine project_id from credentials file: %v", err)
	}

	ctx := context.Background()
	creds, err := google.CredentialsFromJSON(ctx, jsonConfig, datastore.ScopeDatastore)
	if err != nil {
		return nil,"", err
	}
	c,err:=datastore.NewClient(ctx, projectId, option.WithCredentials(creds))
	if err != nil {
		return nil,"", err
	}
	return c,projectId,err
}
