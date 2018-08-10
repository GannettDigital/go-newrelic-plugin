package datastore

import (
	"testing"
	"cloud.google.com/go/datastore"
	"context"
	"time"
	"github.com/franela/goblin"
)



type FakeDataStoreClient struct{
	kinds []DatastoreKind
	err error
}


func NewFakeClient( kinds []DatastoreKind, err error ) Client {
	return Client{
		Dsc: FakeDataStoreClient{
			kinds: kinds,
			err: err,
		},
	}
}


func (fdsc FakeDataStoreClient) GetAll(ctx context.Context, q *datastore.Query, dst interface{}) (keys []*datastore.Key, err error){
	dst = fdsc.kinds
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
					KindName:            "testAsset",
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
				kinds, err := getDatastoreQueryResult(fakeClient.Dsc)
				g.Assert(err).Equal(test.expectedErr)
				g.Assert(kinds).Equal(test.datastoreKind)
			})
		})

	}
}