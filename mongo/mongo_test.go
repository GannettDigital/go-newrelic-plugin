package mongo

import (
	"fmt"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/franela/goblin"
)

var fakeLog = logrus.New()

func TestValidateConfig(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("mongo validateConfig()", func() {
		expected := map[string]struct {
			ExpectedIsNil bool
			MongoConfig   Config
		}{
			"all Fields are set. Host, Password, User, Port, dbName": {true, Config{MongoDBHost: "http://localhost", MongoDBPassword: "Pass", MongoDBUser: "User", MongoDBPort: "80", MongoDB: "Admin"}},
			"no":                         {false, Config{}},
			"Host":                       {false, Config{MongoDBHost: "http://localhost"}},
			"Password":                   {false, Config{MongoDBPassword: "Pass"}},
			"Host, Password, User":       {false, Config{MongoDBHost: "http://localhost", MongoDBPassword: "Pass", MongoDBUser: "User"}},
			"Host, Password, User, Port": {false, Config{MongoDBHost: "http://localhost", MongoDBPassword: "Pass", MongoDBUser: "User", MongoDBPort: "80"}},
		}
		for name, ex := range expected {
			desc := fmt.Sprintf("should return %v when %v fields are set", ex.ExpectedIsNil, name)
			g.It(desc, func() {
				valid := ValidateConfig(ex.MongoConfig)
				g.Assert(valid == nil).Equal(ex.ExpectedIsNil)
			})
		}
	})
}

func TestReadStats(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputLog        *logrus.Logger
		InputSession    Session
		ExpectedRes     []dbStats
		TestDescription string
	}{
		{
			InputLog: logrus.New(),
			InputSession: NewMockSession(
				MockSessionResults{
					DatabaseNamesResult: []string{"foo", "bar"},
				},
				map[string]MockDatabaseResults{
					"foo": MockDatabaseResults{
						RunResult: []byte("{\"DB\":\"foo\",\"Collections\":1,\"Objects\":29,\"AvgObjSize\":1029,\"DataSize\":1024,\"StorageSize\":1020,\"NumExtents\":10,\"Indexes\":100,\"IndexSize\":2048}"),
					},
					"bar": MockDatabaseResults{
						RunResult: []byte("{\"DB\":\"bar\",\"Collections\":2,\"Objects\":30,\"AvgObjSize\":1030,\"DataSize\":1025,\"StorageSize\":1021,\"NumExtents\":11,\"Indexes\":101,\"IndexSize\":2049}"),
					},
				},
			),
			ExpectedRes:     []dbStats{},
			TestDescription: "Should successfully read two database's  stats from mongo",
		},
	}

	for _, test := range tests {
		g.Describe("readDBStats()", func() {
			g.It(test.TestDescription, func() {
				res := readDBStats(test.InputLog, test.InputSession)
				g.Assert(len(res)).Equal(2)
			})
		})
	}
}
