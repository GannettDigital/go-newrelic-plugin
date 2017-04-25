package mongo

import (
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
)

// Session is an interface to access to the Session struct.
type Session interface {
	DB(name string) DataLayer
	DatabaseNames() ([]string, error)
	Run(selector interface{}, update interface{}) error
}

// MongoSession is currently a Mongo session.
type MongoSession struct {
	*mgo.Session
}

// DB shadows *mgo.DB to returns a DataLayer interface instead of *mgo.Database.
func (s MongoSession) DB(name string) DataLayer {
	return &MongoDatabase{Database: s.Session.DB(name)}
}

type DataLayer interface {
	Run(selector interface{}, update interface{}) error
}

// MongoDatabase wraps a mgo.Database to embed methods in models.
type MongoDatabase struct {
	*mgo.Database
}

type MockSession struct {
	SessionResults  MockSessionResults
	DatabaseResults map[string]MockDatabaseResults
	Err             error
}

type MockSessionResults struct {
	DatabaseNamesResult []string
	RunResult           []byte
}

type MockDatabaseResults struct {
	RunResult []byte
}

// DB mocks mgo.Session.DB().
func (fs MockSession) DB(name string) DataLayer {
	mockDatabase := MockDatabase{DatabaseResults: fs.DatabaseResults[name]}
	return mockDatabase
}

// DatabaseNames mocks mgo.Session.DatabaseNames().
func (fs MockSession) DatabaseNames() ([]string, error) {
	return fs.SessionResults.DatabaseNamesResult, nil
}

func (fs MockSession) Run(selector interface{}, update interface{}) error {
	err := json.Unmarshal(fs.SessionResults.RunResult, &update)
	if err != nil {
		fmt.Println(err.Error())
	}
	return fs.Err
}

// MockDatabase satisfies DataLayer and act as a mock.
type MockDatabase struct {
	DatabaseResults MockDatabaseResults
	Err             error
}

// Run mocks mgo.Database(name).Collection(name).
func (db MockDatabase) Run(selector interface{}, update interface{}) error {
	err := json.Unmarshal(db.DatabaseResults.RunResult, &update)
	if err != nil {
		fmt.Println(err.Error())
	}
	return db.Err
}

//Config is the keeper of the config
type Config struct {
	MongoDBUser     string
	MongoDBPassword string
	MongoDBHost     string
	MongoDBPort     string
	MongoDB         string
}

// InventoryData is the data type for inventory data produced by a plugin data
// source and emitted to the agent's inventory data store
type inventoryData map[string]interface{}

// MetricData is the data type for events produced by a plugin data source and
// emitted to the agent's metrics data store
type metricData map[string]interface{}

// EventData is the data type for single shot events
type eventData map[string]interface{}

// PluginData defines the format of the output JSON that plugins will return
type pluginData struct {
	Name            string                   `json:"name"`
	ProtocolVersion string                   `json:"protocol_version"`
	PluginVersion   string                   `json:"plugin_version"`
	Metrics         []metricData             `json:"metrics"`
	Inventory       map[string]inventoryData `json:"inventory"`
	Events          []eventData              `json:"events"`
	Status          string                   `json:"status"`
}

// https://docs.mongodb.com/manual/reference/command/serverStatus/#dbcmd.serverStatus

type serverStatusAsserts struct {
	Regular   int `bson:"regular"`
	Warning   int `bson:"warning"`
	Msg       int `bson:"msg"`
	User      int `bson:"user"`
	Rollovers int `bson:"rollovers"`
}

type serverStatusBackgroundFlushing struct {
	Flushes   int `bson:"flushes"`
	TotalMS   int `bson:"total_ms"`
	AverageMS int `bson:"average_ms"`
	LastMS    int `bson:"last_ms"`
}

type serverStatusConnections struct {
	Current      int   `bson:"current"`
	Available    int   `bson:"available"`
	TotalCreated int64 `bson:"totalCreated"`
}

type serverStatusDur struct {
	Commits            int `bson:"commits"`
	JournaledMB        int `bson:"journaledMB"`
	WriteToDataFilesMB int `bson:"writeToDataFilesMB"`
	Compression        int `bson:"compression"`
	CommitsInWriteLock int `bson:"commitsInWriteLock"`
	EarlyCommits       int `bson:"earlyCommits"`
	TimeMS             struct {
		DT                 int `bson:"dt"`
		PrepLogBuffer      int `bson:"prepLogBuffer"`
		WriteToJournal     int `bson:"writeToJournal"`
		WriteToDataFiles   int `bson:"writeToDataFiles"`
		RemapPrivateView   int `bson:"remapPrivateView"`
		Commits            int `bson:"commits"`
		CommitsInWriteLock int `bson:"commitsInWriteLock"`
	} `bson:"timeMs"`
}

type severStatsExtraInfo struct {
	PageFaults int `bson:"page_faults"`
}

type serverStatsGlobalLock struct {
	TotalTime    int `bson:"totalTime"`
	CurrentQueue struct {
		Total   int `bson:"total"`
		Readers int `bson:"readers"`
		Writers int `bson:"writers"`
	} `bson:"currentQueue"`
	ActiveClients struct {
		Total   int `bson:"total"`
		Readers int `bson:"readers"`
		Writers int `bson:"writers"`
	} `bson:"activeClients"`
}

type serverStatsNetwork struct {
	BytesIn     int64 `bson:"bytesIn"`
	BytesOut    int64 `bson:"bytesOut"`
	NumRequests int64 `bson:"numRequests"`
}

type serverStatsOpcounters struct {
	Insert  int `bson:"insert"`
	Query   int `bson:"query"`
	Update  int `bson:"update"`
	Delete  int `bson:"delete"`
	Getmore int `bson:"getmore"`
	Command int `bson:"command"`
}

type serverStatsStorageEngine struct {
	Name                   string `bson:"name"`
	SupportsCommittedReads bool   `bson:"supportsCommittedReads"`
	Persistent             bool   `bson:"persistent"`
}

type serverStatsWiredTiger struct {
	Cache struct {
		BytesCurrentlyInCache                            int64 `bson:"bytes currently in the cache"`
		FailedEvictionPagesExceedingTheInMemoryMaximumps int64 `bson:"failed eviction of pages that exceeded the in-memory maximum"`
		InMemoryPageSplits                               int   `bson:"in-memory page splits"`
		MaximumBytesConfigured                           int64 `bson:"maximum bytes configured"`
		MaximumPageSizeAtEviction                        int64 `bson:"maximum page size at eviction"`
		ModifiedPagesEvicted                             int   `bson:"modified pages evicted"`
		PagesCurrentlyHeldInTheCache                     int   `bson:"pages currently held in the cache"`
		PagesEvictedByApplicationThreads                 int   `bson:"pages evicted by application threads"`
		PagesEvictedBecauseTheyExeceededTheInMemoryMax   int   `bson:"pages evicted because they exceeded the in-memory maximum"`
		TrackedDirtyBytesInTheCache                      int64 `bson:"tracked dirty bytes in the cache"`
		UnmodifiedPagesEvicted                           int   `bson:"unmodified pages evicted"`
	} `bson:"cache"`
	ConcurrentTransations struct {
		Write struct {
			Out          int `bson:"out"`
			Available    int `bson:"available"`
			TotalTickets int `bson:"totalTickets"`
		} `bson:"write"`
		Read struct {
			Out          int `bson:"out"`
			Available    int `bson:"available"`
			TotalTickets int `bson:"totalTickets"`
		} `bson:"read"`
	} `bson:"concurrentTransactions"`
}

type serverStatsMem struct {
	Bits              int64 `bson:"bits"`
	Resident          int64 `bson:"resident"`
	Virtual           int64 `bson:"virtual"`
	Supported         int64 `bson:"supported"`
	Mapped            int64 `bson:"mapped"`
	MappedWithJournal int64 `bson:"mappedWithJournal"`
}

type serverStatsMetrics struct {
	Cursor struct {
		TimedOut int64 `bson:"timedOut"`
		Open     struct {
			NoTimeout int `bson:"noTimeout"`
			Pinned    int `bson:"pinned"`
			Total     int `bson:"total"`
		} `bson:"open"`
	} `bson:"cursor"`
	Document struct {
		Deleted  int64 `bson:"deleted"`
		Inserted int64 `bson:"inserted"`
		Updated  int64 `bson:"updated"`
		Returned int64 `bson:"returned"`
	} `bson:"document"`
	GetLastError struct {
		Wtimeouts int64 `bson:"wtimeouts"`
		Wtime     struct {
			Num         int64 `bson:"num"`
			TotalMillis int64 `bson:"totalMillis"`
		} `bson:"wtime"`
	} `bson:"getLastError"`
	Operation struct {
		Fastmod        int64 `bson:"fastmod"`
		Idhack         int64 `bson:"idhack"`
		ScanAndOrder   int64 `bson:"scanAndOrder"`
		WriteConflicts int64 `bson:"writeConflicts"`
	} `bson:"operation"`
	QueryExecutor struct {
		Scanned        int64 `bson:"scanned"`
		ScannedObjects int64 `bson:"scannedObjects"`
	} `bson:"queryExecutor"`
}

type serverStatus struct {
	Host               string
	Version            string
	Process            string
	Pid                int
	Uptime             int
	UptimeMillis       int                            `bson:"uptimeMillis"`
	UptimeEstimate     int                            `bson:"uptimeEstimate"`
	Asserts            serverStatusAsserts            `bson:"asserts"`
	BackgroundFlushing serverStatusBackgroundFlushing `bson:"backgroundFlushing"`
	Connections        serverStatusConnections        `bson:"connections"`
	Dur                serverStatusDur                `bson:"dur"`
	ExtraInfo          severStatsExtraInfo            `bson:"extra_info"`
	GlobalLock         serverStatsGlobalLock          `bson:"globalLock"`
	Network            serverStatsNetwork             `bson:"network"`
	OpCounters         serverStatsOpcounters          `bson:"opcounters"`
	OpCountersRepl     serverStatsOpcounters          `bson:"opcountersRepl"`
	StorageEngine      serverStatsStorageEngine       `bson:"storageEngine"`
	WiredTiger         serverStatsWiredTiger          `bson:"wiredTiger"`
	Mem                serverStatsMem                 `bson:"mem"`
	Metrics            serverStatsMetrics             `bson:"metrics"`
}

type dbStats struct {
	DB          string `bson:"db"`
	Collections int    `bson:"collections"`
	Objects     int    `bson:"objects"`
	AvgObjSize  int64  `bson:"avgObjSize"`
	DataSize    int64  `bson:"dataSize"`
	StorageSize int64  `bson:"storageSize"`
	NumExtents  int    `bson:"numExtents"`
	Indexes     int    `bson:"indexes"`
	IndexSize   int64  `bson:"indexSize"`
}

type ReplStats struct {
	Set                     string       `bson:"set" json:"set"`
	Date                    time.Time    `bson:"date" json:"date"`
	MyState                 int          `bson:"myState" json:"myState"`
	Term                    int64        `bson:"term" json:"term"`
	HeartbeatIntervalMillis int64        `bson:"heartbeatIntervalMillis" json:"heartbeatIntervalMillis"`
	Members                 []ReplMember `bson:"members" json:"members"`
	OK                      int          `bson:"ok" json:"ok"`
}

type ReplMember struct {
	ID                int        `bson:"_id" json:"_id"`
	Name              string     `bson:"name" json:"name"`
	Health            int        `bson:"health" json:"health"`
	State             int        `bson:"state" json:"state"`
	StateStr          string     `bson:"stateStr" json:"stateStr"`
	Uptime            int64      `bson:"uptime" json:"uptime"`
	Optime            ReplOptime `bson:"optime" json:"optime"`
	OptimeDate        time.Time  `bson:"optimeDate" json:"optimeDate"`
	ElectionTime      int64      `bson:"electionTime" json:"electionTime"`
	ElectionDate      time.Time  `bson:"electionDate" json:"electionDate"`
	LastHeartbeat     time.Time  `bson:"lastHeartbeat" json:"lastHeartbeat"`
	LastHeartbeatRecv time.Time  `bson:"lastHeartbeatRecv" json:"lastHeartbeatRecv"`
	PingMS            int64      `bson:"pingMs" json:"pingMs"`
	SyncingTo         string     `bson:"syncingTo" json:"syncingTo"`
	ConfigVersion     int        `bson:"configVersion" json:"configVersion"`
	Self              bool       `bson:"self" json:"self"`
}

type ReplOptime struct {
	TS int64 `bson:"ts" json:"ts"`
	T  int64 `bson:"t" json:"t"`
}
