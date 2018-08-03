package datastore

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/monitoring/v3"
	"github.com/Sirupsen/logrus"
	"cloud.google.com/go/datastore"
	"google.golang.org/api/option"
)

var metrics = []string{
	"datastore.googleapis.com/api/request_count",
	"datastore.googleapis.com/entity/read_sizes",
	"datastore.googleapis.com/entity/write_sizes",
	"datastore.googleapis.com/index/write_count",
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

func Run(log *logrus.Logger, prettyPrint bool, version string) {
	stackdriverMetrics()
	datastoreStats()
}

func datastoreStats(){
	ctx := context.Background()
	dsClient, err := datastore.NewClient(ctx,"gannett-api-services-stage",option.WithCredentialsFile(`/Users/jstorer/Downloads/gannett-api-services-stage-76507247423e.json`))
	if err != nil{
		log.Fatal("Error connecting to datastore")
	}


	q := datastore.NewQuery("__Stat_Kind__").Order("kind_name")
	kinds := []*DatastoreKind{}
	_, err = dsClient.GetAll(ctx,q,&kinds)
	if err != nil{
		log.Fatal(err)
	}

	for _, k := range kinds{
		fmt.Printf("\nkind %q\t%d entries\t%d bytes\n", k.KindName, k.Count, k.Bytes)
	}
}

func stackdriverMetrics(){
	ctx := context.Background()

	hc, err := google.DefaultClient(ctx, monitoring.MonitoringScope)

	if err != nil {
		log.Fatal(err)
	}

	s, err := monitoring.New(hc)

	if err != nil {
		log.Fatal(err)
	}


	projectID := "gannett-api-services-stage"

	//loop through datastore metrics
	for _, metric := range metrics {
		if err := listMetricDescriptors(s, projectID, metric); err != nil {
			log.Fatal(err)
		}
		if err := listTimeSeries(s, projectID, metric); err != nil {
			log.Fatal(err)
		}
	}
}


func listMetricDescriptors(s *monitoring.Service, projectID string, metric string) error {

	resp, err := s.Projects.MetricDescriptors.List(projectResource(projectID)).
		Filter(fmt.Sprintf("metric.type=%q", metric)).
		Do()
	if err != nil {
		return fmt.Errorf("Could not list metric descriptors: %v", err)
	}

	log.Printf("listMetricDescriptors %s\n", formatResource(resp))
	return nil
}

func listTimeSeries(s *monitoring.Service, projectID string, metric string) error {
	startTime := time.Now().UTC().Add(-time.Hour)
	endTime := startTime.Add(5 * time.Minute)

	resp, err := s.Projects.TimeSeries.List(projectResource(projectID)).
		PageSize(3).
		Filter(fmt.Sprintf("metric.type=\"%s\"", metric)).
		IntervalStartTime(startTime.Format(time.RFC3339)).
		IntervalEndTime(endTime.Format(time.RFC3339)).
		Do()
	if err != nil {
		return fmt.Errorf("Could not list time series: %v", err)
	}

	log.Printf("listTimeseries %s\n", formatResource(resp))
	return nil
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
