package datastore

import (
	"context"
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/franela/goblin"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/monitoring/v3"
)

type FakeDataStoreClient struct {
	kindsFake []DatastoreKind
	err       error
}

func NewFakeClient(kindsInit []DatastoreKind, err error) ClientDatastore {
	return ClientDatastore{
		Dsc: FakeDataStoreClient{
			kindsFake: kindsInit,
			err:       err,
		},
	}
}

func (fdsc FakeDataStoreClient) GetAll(ctx context.Context, q *datastore.Query, dst interface{}) (keys []*datastore.Key, err error) {

	dv := reflect.ValueOf(dst).Elem()

	for _, value := range fdsc.kindsFake {
		tempDataStore := DatastoreKind{
			BuiltinIndexBytes:   value.BuiltinIndexBytes,
			BuiltinIndexCount:   value.BuiltinIndexCount,
			CompositeIndexBytes: value.CompositeIndexBytes,
			CompositeIndexCount: value.CompositeIndexCount,
			EntityBytes:         value.EntityBytes,
			Bytes:               value.Bytes,
			Count:               value.Count,
			KindName:            value.KindName,
			Timestamp:           value.Timestamp,
		}

		dv.Set(reflect.Append(dv, reflect.ValueOf(tempDataStore)))
	}

	return []*datastore.Key{}, err
}

func TestDatastoreStatKindQueryResult(t *testing.T) {
	g := goblin.Goblin(t)
	var tests = []struct {
		datastoreKind []DatastoreKind
		err           error
		expectedErr   error
		description   string
	}{
		{
			[]DatastoreKind{
				{
					BuiltinIndexBytes:   1,
					BuiltinIndexCount:   2,
					CompositeIndexBytes: 3,
					CompositeIndexCount: 4,
					EntityBytes:         5,
					Bytes:               6,
					Count:               7,
					KindName:            "testAsset1",
					Timestamp:           time.Now(),
				},
			},
			nil,
			nil,
			"valid return one kind",
		},
		{
			[]DatastoreKind{
				{
					BuiltinIndexBytes:   1,
					BuiltinIndexCount:   2,
					CompositeIndexBytes: 3,
					CompositeIndexCount: 4,
					EntityBytes:         5,
					Bytes:               6,
					Count:               7,
					KindName:            "testAsset1",
					Timestamp:           time.Unix(123, 0),
				},
				{
					BuiltinIndexBytes:   2,
					BuiltinIndexCount:   3,
					CompositeIndexBytes: 4,
					CompositeIndexCount: 5,
					EntityBytes:         6,
					Bytes:               7,
					Count:               8,
					KindName:            "testAsset2",
					Timestamp:           time.Unix(123, 123),
				},
			},
			nil,
			nil,
			"valid return multiple kinds",
		},
	}

	for _, test := range tests {
		g.Describe("DatastoreStatKindQueryResult()", func() {
			g.It(test.description, func() {
				fakeClient := NewFakeClient(test.datastoreKind, test.err)
				kindsResult, err := fakeClient.KindStats()
				g.Assert(err).Equal(test.expectedErr)
				g.Assert(kindsResult).Equal(test.datastoreKind)
			})
		})

	}
}

func TestDataStoreData(t *testing.T) {
	g := goblin.Goblin(t)
	var tests = []struct {
		projectID           string
		datastoreKind       []DatastoreKind
		datastoreDataWanted []map[string]interface{}
		description         string
	}{
		{
			"testProjectID",
			[]DatastoreKind{
				{
					BuiltinIndexBytes:   1,
					BuiltinIndexCount:   2,
					CompositeIndexBytes: 3,
					CompositeIndexCount: 4,
					EntityBytes:         5,
					Bytes:               6,
					Count:               7,
					KindName:            "testAsset1",
					Timestamp:           time.Now(),
				},
			},
			[]map[string]interface{}{
				{
					"event_type":                         "DatastoreSample",
					"provider":                           "datastoreQuery",
					"datastoreQuery.builtinIndexBytes":   1,
					"datastoreQuery.builtinIndexCount":   2,
					"datastoreQuery.compositeIndexBytes": 3,
					"datastoreQuery.compositeIndexCount": 4,
					"datastoreQuery.entityBytes":         5,
					"datastoreQuery.bytes":               6,
					"datastoreQuery.count":               7,
					"datastoreQuery.kindName":            "testAsset1",
					"datastoreQuery.projectId":           "testProjectID",
					"datastoreQuery.timestamp":           time.Now().Unix(),
				},
			},
			"valid return one kind",
		},
		{
			"testProjectID",
			[]DatastoreKind{
				{
					BuiltinIndexBytes:   1,
					BuiltinIndexCount:   2,
					CompositeIndexBytes: 3,
					CompositeIndexCount: 4,
					EntityBytes:         5,
					Bytes:               6,
					Count:               7,
					KindName:            "testAsset1",
					Timestamp:           time.Now(),
				},
				{
					BuiltinIndexBytes:   2,
					BuiltinIndexCount:   3,
					CompositeIndexBytes: 4,
					CompositeIndexCount: 5,
					EntityBytes:         6,
					Bytes:               7,
					Count:               8,
					KindName:            "testAsset2",
					Timestamp:           time.Now(),
				},
			},
			[]map[string]interface{}{
				{
					"event_type":                         "DatastoreSample",
					"provider":                           "datastoreQuery",
					"datastoreQuery.builtinIndexBytes":   1,
					"datastoreQuery.builtinIndexCount":   2,
					"datastoreQuery.compositeIndexBytes": 3,
					"datastoreQuery.compositeIndexCount": 4,
					"datastoreQuery.entityBytes":         5,
					"datastoreQuery.bytes":               6,
					"datastoreQuery.count":               7,
					"datastoreQuery.kindName":            "testAsset1",
					"datastoreQuery.projectId":           "testProjectID",
					"datastoreQuery.timestamp":           time.Now().Unix(),
				},
				{
					"event_type":                         "DatastoreSample",
					"provider":                           "datastoreQuery",
					"datastoreQuery.builtinIndexBytes":   2,
					"datastoreQuery.builtinIndexCount":   3,
					"datastoreQuery.compositeIndexBytes": 4,
					"datastoreQuery.compositeIndexCount": 5,
					"datastoreQuery.entityBytes":         6,
					"datastoreQuery.bytes":               7,
					"datastoreQuery.count":               8,
					"datastoreQuery.kindName":            "testAsset2",
					"datastoreQuery.projectId":           "testProjectID",
					"datastoreQuery.timestamp":           time.Now().Unix(),
				},
			},
			"valid return multiple kinds",
		},
	}
	for _, test := range tests {
		g.Describe("DatastoreData()", func() {
			g.It(test.description, func() {
				fakeClient := NewFakeClient(test.datastoreKind, nil)
				fakeClient.projectId = test.projectID
				data := fakeClient.DatastoreData(test.datastoreKind)
				g.Assert(data).Equal(test.datastoreDataWanted)
			})
		})

	}
}

func TestStackdriverData(t *testing.T) {
	g := goblin.Goblin(t)
	var tests = []struct {
		resp        monitoring.ListTimeSeriesResponse
		dataWanted  []map[string]interface{}
		errWanted   error
		description string
	}{
		{
			monitoring.ListTimeSeriesResponse{
				ExecutionErrors: []*monitoring.Status{},
				NextPageToken:   "",
				TimeSeries: []*monitoring.TimeSeries{
					{
						Metric: &monitoring.Metric{
							Labels: map[string]string{
								"api_method":    "Commit",
								"response_code": "OK",
							},
							Type: "datastore.googleapis.com/api/request_count",
						},
						MetricKind: "DELTA",
						Points: []*monitoring.Point{
							{
								Interval: &monitoring.TimeInterval{
									EndTime:   "2018-08-10T20:17:37.754Z",
									StartTime: "2018-08-10T20:16:37.754Z",
								},
								Value: &monitoring.TypedValue{
									Int64Value: googleapi.Int64(2),
								},
							},
						},
						Resource: &monitoring.MonitoredResource{
							Labels: map[string]string{
								"module_id":  "__unknown__",
								"project_id": "gannett-api-services-stage",
								"version_id": "__unknown__",
							},
							Type: "datastore_request",
						},
						ValueType: "INT64",
					},
				},
			},
			[]map[string]interface{}{
				{
					"event_type":                        "DatastoreSample",
					"provider":                          "datastoreStackdriver",
					"datastoreStackdriver.apiMethod":    "Commit",
					"datastoreStackdriver.responseCode": "OK",
					"datastoreStackdriver.metricType":   "datastore.googleapis.com/api/request_count",
					"datastoreStackdriver.metricKind":   "DELTA",
					"datastoreStackdriver.timestamp":    int64(1533932197),
					"datastoreStackdriver.value":        int64(2),
					"datastoreStackdriver.projectId":    "gannett-api-services-stage",
					"datastoreStackdriver.resourceType": "datastore_request",
					"datastoreStackdriver.valueType":    "INT64",
				},
			},
			nil,
			"One timeseries point",
		},
		{
			monitoring.ListTimeSeriesResponse{
				ExecutionErrors: []*monitoring.Status{},
				NextPageToken:   "",
				TimeSeries: []*monitoring.TimeSeries{
					{
						Metric: &monitoring.Metric{
							Labels: map[string]string{
								"api_method":    "Commit",
								"response_code": "OK",
							},
							Type: "datastore.googleapis.com/api/request_count",
						},
						MetricKind: "DELTA",
						Points: []*monitoring.Point{
							{
								Interval: &monitoring.TimeInterval{
									EndTime:   "2018-08-10T20:17:37.754Z",
									StartTime: "2018-08-10T20:16:37.754Z",
								},
								Value: &monitoring.TypedValue{
									Int64Value: googleapi.Int64(2),
								},
							},
						},
						Resource: &monitoring.MonitoredResource{
							Labels: map[string]string{
								"module_id":  "__unknown__",
								"project_id": "gannett-api-services-stage",
								"version_id": "__unknown__",
							},
							Type: "datastore_request",
						},
						ValueType: "INT64",
					},
					{
						Metric: &monitoring.Metric{
							Labels: map[string]string{
								"api_method":    "Query",
								"response_code": "OK",
							},
							Type: "datastore.googleapis.com/api/request_count",
						},
						MetricKind: "DELTA",
						Points: []*monitoring.Point{
							{
								Interval: &monitoring.TimeInterval{
									EndTime:   "2018-08-10T20:17:37.754Z",
									StartTime: "2018-08-10T20:16:37.754Z",
								},
								Value: &monitoring.TypedValue{
									Int64Value: googleapi.Int64(2),
								},
							},
						},
						Resource: &monitoring.MonitoredResource{
							Labels: map[string]string{
								"module_id":  "__unknown__",
								"project_id": "gannett-api-services-stage",
								"version_id": "__unknown__",
							},
							Type: "datastore_request",
						},
						ValueType: "INT64",
					},
				},
			},
			[]map[string]interface{}{
				{
					"event_type":                        "DatastoreSample",
					"provider":                          "datastoreStackdriver",
					"datastoreStackdriver.apiMethod":    "Commit",
					"datastoreStackdriver.responseCode": "OK",
					"datastoreStackdriver.metricType":   "datastore.googleapis.com/api/request_count",
					"datastoreStackdriver.metricKind":   "DELTA",
					"datastoreStackdriver.timestamp":    int64(1533932197),
					"datastoreStackdriver.value":        int64(2),
					"datastoreStackdriver.projectId":    "gannett-api-services-stage",
					"datastoreStackdriver.resourceType": "datastore_request",
					"datastoreStackdriver.valueType":    "INT64",
				},
				{
					"event_type":                        "DatastoreSample",
					"provider":                          "datastoreStackdriver",
					"datastoreStackdriver.apiMethod":    "Query",
					"datastoreStackdriver.responseCode": "OK",
					"datastoreStackdriver.metricType":   "datastore.googleapis.com/api/request_count",
					"datastoreStackdriver.metricKind":   "DELTA",
					"datastoreStackdriver.timestamp":    int64(1533932197),
					"datastoreStackdriver.value":        int64(2),
					"datastoreStackdriver.projectId":    "gannett-api-services-stage",
					"datastoreStackdriver.resourceType": "datastore_request",
					"datastoreStackdriver.valueType":    "INT64",
				},
			},
			nil,
			"Two timeseries points",
		},
		{
			monitoring.ListTimeSeriesResponse{
				ExecutionErrors: []*monitoring.Status{},
				NextPageToken:   "",
				TimeSeries: []*monitoring.TimeSeries{
					{
						Metric: &monitoring.Metric{
							Labels: map[string]string{
								"op": "CREATE",
							},
							Type: "datastore.googleapis.com/entity/write_sizes",
						},
						MetricKind: "DELTA",
						Points: []*monitoring.Point{
							{
								Interval: &monitoring.TimeInterval{
									EndTime:   "2018-08-10T20:17:37.754Z",
									StartTime: "2018-08-10T20:16:37.754Z",
								},
								Value: &monitoring.TypedValue{
									DistributionValue: &monitoring.Distribution{
										BucketCounts: googleapi.Int64s{0, 1, 1},
										Count:        2,
										Mean:         100,
									},
								},
							},
						},
						Resource: &monitoring.MonitoredResource{
							Labels: map[string]string{
								"module_id":  "__unknown__",
								"project_id": "gannett-api-services-stage",
								"version_id": "__unknown__",
							},
							Type: "datastore_request",
						},
						ValueType: "DISTRIBUTION",
					},
				},
			},
			[]map[string]interface{}{
				{
					"event_type":                      "DatastoreSample",
					"provider":                        "datastoreStackdriver",
					"datastoreStackdriver.op":         "CREATE",
					"datastoreStackdriver.mean":       float64(100),
					"datastoreStackdriver.type":       "",
					"datastoreStackdriver.metricType": "datastore.googleapis.com/entity/write_sizes",
					"datastoreStackdriver.timestamp":  int64(1533932197),
					"datastoreStackdriver.projectId":  "gannett-api-services-stage",
					"datastoreStackdriver.valueType":  "DISTRIBUTION",
				},
				{
					"event_type":                        "DatastoreSample",
					"provider":                          "datastoreStackdriver",
					"datastoreStackdriver.type":         "",
					"datastoreStackdriver.op":           "CREATE",
					"datastoreStackdriver.metricType":   "datastore.googleapis.com/entity/write_sizes",
					"datastoreStackdriver.metricKind":   "DELTA",
					"datastoreStackdriver.timestamp":    int64(1533932197),
					"datastoreStackdriver.bucket":       float64(4),
					"datastoreStackdriver.projectId":    "gannett-api-services-stage",
					"datastoreStackdriver.resourceType": "datastore_request",
					"datastoreStackdriver.valueType":    "DISTRIBUTION",
				},
				{
					"event_type":                        "DatastoreSample",
					"provider":                          "datastoreStackdriver",
					"datastoreStackdriver.op":           "CREATE",
					"datastoreStackdriver.type":         "",
					"datastoreStackdriver.bucket":       float64(16),
					"datastoreStackdriver.metricType":   "datastore.googleapis.com/entity/write_sizes",
					"datastoreStackdriver.metricKind":   "DELTA",
					"datastoreStackdriver.timestamp":    int64(1533932197),
					"datastoreStackdriver.projectId":    "gannett-api-services-stage",
					"datastoreStackdriver.valueType":    "DISTRIBUTION",
					"datastoreStackdriver.resourceType": "datastore_request",
				},
			},
			nil,
			"distribution point",
		},
	}

	for _, test := range tests {
		g.Describe("StackdriverData()", func() {
			g.It(test.description, func() {
				data, err := StackdriverData(&test.resp)
				for id := range data {
					g.Assert(data[id]).Equal(test.dataWanted[id])

				}
				g.Assert(err).Equal(test.errWanted)
			})
		})

	}

}
