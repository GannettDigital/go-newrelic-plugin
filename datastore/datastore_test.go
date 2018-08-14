package datastore

import (
	"context"
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/franela/goblin"
)

type FakeDataStoreClient struct {
	kindsFake []DatastoreKind
	err       error
}

func NewFakeClient(kindsInit []DatastoreKind, err error) Client {
	return Client{
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

func TestGetDatastoreQueryResult(t *testing.T) {
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
		g.Describe("getDatastoreQueryResult()", func() {
			g.It(test.description, func() {
				fakeClient := NewFakeClient(test.datastoreKind, test.err)
				kindsResult, err := getDatastoreQueryResult(fakeClient.Dsc)
				g.Assert(err).Equal(test.expectedErr)
				g.Assert(kindsResult).Equal(test.datastoreKind)
			})
		})

	}
}

func TestGetDataStoreData(t *testing.T) {
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
		g.Describe("getDatastoreData()", func() {
			g.It(test.description, func() {
				data := getDatastoreData(test.datastoreKind, test.projectID)
				g.Assert(data).Equal(test.datastoreDataWanted)
			})
		})

	}
}


