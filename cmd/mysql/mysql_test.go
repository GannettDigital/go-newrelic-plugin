package mysql

import (
	//	"fmt"
	//	"net"
	//	"os"
	//	"strconv"
	"testing"

	"github.com/franela/goblin"
	"github.com/Sirupsen/logrus"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	log = logrus.New()
	config = mysqlConfig{
		host:     "HOST",
		port:     "PORT",
		user:     "USER",
		password: "PASSWORD",
		database: "DATABASE",
		queries:  "show status; show global variables;",
		prefixes: "galera_ innodb_ net_ performance_ Galera_ Innodb_ Net_ Performance_",
	}
}

// func getMetrics(db *sql.DB) (map[string]interface{}, error) {
func TestGetMetrics(t *testing.T) {
	// Setup mock db
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"Variable_name", "Value"}).
		AddRow("Aborted_clients", 0)
	mock.ExpectQuery("^show status$").WillReturnRows(rows)

	g := goblin.Goblin(t)
	var tests = []struct {
		TestDescription string
		result          map[string]interface{}
	}{
		{
			TestDescription: "Should get status without error",
			result:          map[string]interface{}{"event_type": "DatastoreSample", "provider": "mysql", "mysql.AbortedClients": 0},
		},
	}
	for _, test := range tests {
		g.Describe("getMetrics)", func() {
			g.It(test.TestDescription, func() {
				name, _ := getMetrics(db)
				g.Assert(name).Equal(test.result)
			})
		})
	}

	getMetrics(db)
}

func TestMetricName(t *testing.T) {
	g := goblin.Goblin(t)
	var tests = []struct {
		TestDescription string
		metric          string
		result          string
	}{
		{
			TestDescription: "Should get metric name without error",
			metric:          "galera_wsrep_cluster_size",
			result:          "mysql.galera.wsrepClusterSize",
		},
	}
	for _, test := range tests {
		g.Describe("getMetricName)", func() {
			g.It(test.TestDescription, func() {
				name := metricName(test.metric)
				g.Assert(name).Equal(test.result)
			})
		})
	}
}

func TestCamelCase(t *testing.T) {
	g := goblin.Goblin(t)
	var tests = []struct {
		TestDescription string
		src             string
		result          string
	}{
		{
			TestDescription: "Should convert to camel case without error",
			src:             "one_two_three_four",
			result:          "oneTwoThreeFour",
		},
		{
			TestDescription: "Should convert to camel case with colon without error",
			src:             "one:two_three_four",
			result:          "one.twoThreeFour",
		},
	}
	for _, test := range tests {
		g.Describe("camelCase)", func() {
			g.It(test.TestDescription, func() {
				camel := camelCase(test.src)
				g.Assert(camel).Equal(test.result)
			})
		})
	}
}

func TestAsValue(t *testing.T) {
	g := goblin.Goblin(t)
	var tests = []struct {
		TestDescription string
		value           string
		result          interface{}
	}{
		{
			TestDescription: "Should convert string to string without error",
			value:           "string",
			result:          "string",
		},
		{
			TestDescription: "Should convert string to int without error",
			value:           "1",
			result:          1,
		},
		{
			TestDescription: "Should convert string to float without error",
			value:           "1.0",
			result:          1.0,
		},
		{
			TestDescription: "Should convert string to boolean without error",
			value:           "true",
			result:          true,
		},
	}
	for _, test := range tests {
		g.Describe("asValue)", func() {
			g.It(test.TestDescription, func() {
				value := asValue(test.value)
				g.Assert(value).Equal(test.result)
			})
		})
	}
}
func TestFixPrefix(t *testing.T) {
	g := goblin.Goblin(t)
	var tests = []struct {
		TestDescription string
		metric          string
		result          string
	}{
		{
			TestDescription: "Should get metric name without error",
			metric:          "Innodb_buffer_pool_dump_status",
			result:          "mysql.Innodb.bufferPoolDumpStatus",
		},
	}
	for _, test := range tests {
		g.Describe("fixPrefix)", func() {
			g.It(test.TestDescription, func() {
				name := metricName(test.metric)
				g.Assert(name).Equal(test.result)
			})
		})
	}

}
func TestGenerateDSN(t *testing.T) {
	g := goblin.Goblin(t)
	var tests = []struct {
		TestDescription string
		result          string
	}{
		{
			TestDescription: "Should generate DSN without error",
			result:          "USER:PASSWORD@tcp(HOST:PORT)/DATABASE",
		},
	}
	for _, test := range tests {
		g.Describe("generateDSN)", func() {
			g.It(test.TestDescription, func() {
				dsn := generateDSN()
				g.Assert(dsn).Equal(test.result)
			})
		})
	}
}
