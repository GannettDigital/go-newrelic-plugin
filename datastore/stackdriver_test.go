package datastore

import (
	"testing"
	"github.com/franela/goblin"
	"google.golang.org/api/monitoring/v3"
	"google.golang.org/api/googleapi"
)

func TestGetStackdriverData(t *testing.T) {
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
				},
			},
			nil,
			"Two timeseries points",
		},
	}

	for _, test := range tests {
		g.Describe("stackdriverData()", func() {
			g.It(test.description, func() {
				data, err := stackdriverData(&test.resp)
				for id, _ := range data {
					g.Assert(data[id]).Equal(test.dataWanted[id])

				}
				g.Assert(err).Equal(test.errWanted)
			})
		})

	}

}