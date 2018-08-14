package datastore

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/buger/jsonparser"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/monitoring/v3"
)

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

//stackdriverResp gets the data of the wanted metric from the stackdriver API
func stackdriverResp(projectId string, metric string) (*monitoring.ListTimeSeriesResponse, error) {
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

//stackdriverData converts a ListTimeSeriesResponse to a []map[string]interface{} to be used for the final output to be sent to new relic
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

