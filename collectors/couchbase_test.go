package collectors

import (
	"errors"
	"reflect"
	"testing"

	fake "github.com/GannettDigital/go-newrelic-plugin/collectors/fake"
	"github.com/franela/goblin"
)

var couchbaseFakeConfig CouchbaseConfig

func init() {
	couchbaseFakeConfig = CouchbaseConfig{
		CouchbaseHost:     "http://localhost",
		CouchbasePassword: "secure",
		CouchbasePort:     "15672",
		CouchbaseUser:     "admin",
	}
}

func TestGetAllBucketsInfo(t *testing.T) {
	g := goblin.Goblin(t)

	type AllBucketsInfoTestResults struct {
		ErrorShouldBeNil      bool
		ClusterInfoShouldHave int
	}
	var tests = []struct {
		HTTPRunner      fake.HTTPResult
		TestResults     AllBucketsInfoTestResults
		TestDescription string
	}{
		{
			HTTPRunner: fake.HTTPResult{
				Code: 200,
				Data: []byte("[{\"name\":\"test1\",\"uri\":\"/pools/default/buckets/test1\",\"stats\":{\"uri\":\"/pools/default/buckets/test1/stats\"}},{\"name\":\"test2\",\"uri\":\"/pools/default/buckets/test2\",\"stats\":{\"uri\":\"/pools/default/buckets/test2/stats\"}}]"),
			},
			TestResults: AllBucketsInfoTestResults{
				ErrorShouldBeNil:      true,
				ClusterInfoShouldHave: 2,
			},
			TestDescription: "Successfully GET List of buckets",
		},
		{
			HTTPRunner: fake.HTTPResult{
				Code: 500,
				Data: []byte("[{\"name\":\"test1\",\"uri\":\"/pools/default/buckets/test1\",\"stats\":{\"uri\":\"/pools/default/buckets/test1/stats\"}},{\"name\":\"test2\",\"uri\":\"/pools/default/buckets/test2\",\"stats\":{\"uri\":\"/pools/default/buckets/test2/stats\"}}]"),
				Err:  errors.New("Internal Error"),
			},
			TestResults: AllBucketsInfoTestResults{
				ErrorShouldBeNil:      false,
				ClusterInfoShouldHave: 0,
			},
			TestDescription: "Fail GET List of buckets",
		},
	}

	for _, test := range tests {
		g.Describe("getAllBucketsInfo()", func() {
			g.It(test.TestDescription, func() {
				runner = test.HTTPRunner
				result, err := getAllBucketsInfo(couchbaseFakeConfig)
				g.Assert(err == nil).Equal(test.TestResults.ErrorShouldBeNil)
				g.Assert(reflect.DeepEqual(len(result), test.TestResults.ClusterInfoShouldHave)).Equal(true)
			})
		})
	}
}

func TestAvgIntSample(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		TestDescription string
		TestData        []int
		ExpectedResult  float32
	}{
		{
			TestDescription: "Successfully Avg int [0, 100, 100]",
			TestData:        []int{0, 100, 100},
			ExpectedResult:  66.666664,
		},
		{
			TestDescription: "Successfully Avg int []",
			TestData:        []int{},
			ExpectedResult:  0,
		},
	}

	for _, test := range tests {
		g.Describe("avgIntSample()", func() {
			g.It(test.TestDescription, func() {
				result := avgIntSample(test.TestData)
				g.Assert(reflect.DeepEqual(result, test.ExpectedResult)).Equal(true)
			})
		})
	}
}

func TestAvgInt64Sample(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		TestDescription string
		TestData        []int64
		ExpectedResult  float32
	}{
		{
			TestDescription: "Successfully Avg int64 [355602721792, 355602721792, 355602721792]",
			TestData:        []int64{355602721792, 355602721792, 355602721792},
			ExpectedResult:  3.5560273e+11,
		},
		{
			TestDescription: "Successfully Avg int64 []",
			TestData:        []int64{},
			ExpectedResult:  0,
		},
	}

	for _, test := range tests {
		g.Describe("avgInt64Sample()", func() {
			g.It(test.TestDescription, func() {
				result := avgInt64Sample(test.TestData)
				g.Assert(reflect.DeepEqual(result, test.ExpectedResult)).Equal(true)
			})
		})
	}
}

func TestAvgFloat32Sample(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		TestDescription string
		TestData        []float32
		ExpectedResult  float32
	}{
		{
			TestDescription: "Successfully Avg float32 [100.00, 250.33, 98.12]",
			TestData:        []float32{100.00, 250.33, 98.12},
			ExpectedResult:  149.48334,
		},
		{
			TestDescription: "Successfully Avg int []",
			TestData:        []float32{},
			ExpectedResult:  0,
		},
	}

	for _, test := range tests {
		g.Describe("avgFloat32Sample()", func() {
			g.It(test.TestDescription, func() {
				result := avgFloat32Sample(test.TestData)
				g.Assert(reflect.DeepEqual(result, test.ExpectedResult)).Equal(true)
			})
		})
	}
}

func TestFormatBucketInfoStructToMap(t *testing.T) {
	g := goblin.Goblin(t)

	var couchbucketStats CouchbaseBucketStats
	couchbucketStats.OP.Samples.GetHits = []int{2, 2, 2}
	var tests = []struct {
		TestDescription string
		TestData        CompleteBucketInfo
		ExpectedResult  map[string]interface{}
	}{
		{
			TestDescription: "Successfully Convert Complete Bucket info to Map Interface",
			TestData: CompleteBucketInfo{
				bucketInfo: CouchbaseBucketStatsUri{
					Name: "TestName",
				},
				bucketStats: couchbucketStats,
			},
			ExpectedResult: map[string]interface{}{
				"couchbase.by_bucket.name":     "TestName",
				"couchbase.by_bucket.get_hits": float32(2),
			},
		},
	}

	for _, test := range tests {
		g.Describe("formatBucketInfoStructToMap()", func() {
			g.It(test.TestDescription, func() {
				result := formatBucketInfoStructToMap(test.TestData)
				g.Assert(result["couchbase.by_bucket.name"] == test.ExpectedResult["couchbase.by_bucket.name"]).Equal(true)
				g.Assert(result["couchbase.by_bucket.get_hits"] == test.ExpectedResult["couchbase.by_bucket.get_hits"]).Equal(true)
			})
		})
	}
}
