package datastore

import (
	"testing"
		"cloud.google.com/go/datastore"
	"context"
	)



type FakeDataStoreClient struct{
	kinds []*DatastoreKind
	err error
}


func NewFakeClient( kinds []*DatastoreKind, err error ) Client {
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

	// todo.. fill out []DatastoreKind and then validate it
	fakeClient := NewFakeClient( []*DatastoreKind{

	},
	nil)


	// now, we can use this fake Client to pass into getDatastoreQueryResult as it will hit our mock GetAll above
	getDatastoreQueryResult(fakeClient.Dsc)
}