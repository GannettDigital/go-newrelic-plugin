package datastore

import (
	"testing"
	"cloud.google.com/go/datastore"
	"context"
	"time"
	"github.com/franela/goblin"
	"reflect"
	)



type FakeDataStoreClient struct{
	kindsFake []DatastoreKind
	err error
}


func NewFakeClient( kindsInit []DatastoreKind, err error ) Client {
	return Client{
		Dsc: FakeDataStoreClient{
			kindsFake: kindsInit,
			err: err,
		},
	}
}


func (fdsc FakeDataStoreClient) GetAll(ctx context.Context, q *datastore.Query, dst interface{}) (keys []*datastore.Key, err error){

	dv  := reflect.ValueOf(dst).Elem()

	for _, value := range fdsc.kindsFake {
		tempDataStore := DatastoreKind{
			BuiltinIndexBytes: value.BuiltinIndexBytes,
			// todo fill out rest
		}

		dv.Set(reflect.Append(dv, reflect.ValueOf(tempDataStore)))
	}


	return []*datastore.Key{}, err
}


func TestGetDatastoreQueryResult(t *testing.T){
	g := goblin.Goblin(t)
	var tests = []struct{
		datastoreKind []DatastoreKind
		err error
		expectedErr error
		description string
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
			"valid return one type data",
		},
	}

	for _, test := range tests{
		g.Describe("getDatastoreQueryResult()", func() {
			g.It(test.description, func() {
				fakeClient := NewFakeClient(test.datastoreKind,test.err)
				kindsResult, err := getDatastoreQueryResult(fakeClient.Dsc)
				g.Assert(err).Equal(test.expectedErr)
				g.Assert(kindsResult).Equal(test.datastoreKind)
			})
		})

	}
}