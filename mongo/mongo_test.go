package mongo

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/franela/goblin"
)

var fakeLog = logrus.New()

func TestRun(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputLog        *logrus.Logger
		InputSession    Session
		InputConfig     Config
		InputPretty     bool
		InputVersion    string
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
			InputConfig: Config{
				MongoDBUser:     "User",
				MongoDBPassword: "Password",
				MongoDBHost:     "localhost",
				MongoDBPort:     "1234",
				MongoDB:         "admin",
			},
			InputPretty:     false,
			InputVersion:    "0.0.1",
			TestDescription: "Should successfully perform a run without error",
		},
	}

	for _, test := range tests {
		g.Describe("Run()", func() {
			g.It(test.TestDescription, func() {
				Run(test.InputLog, test.InputSession, test.InputConfig, test.InputPretty, test.InputVersion)
				g.Assert(true).IsTrue()
			})
		})
	}
}

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
			ExpectedRes: []dbStats{
				dbStats{DB: "foo", Collections: 1, Objects: 29, AvgObjSize: 1029, DataSize: 1024, StorageSize: 1020, NumExtents: 10, Indexes: 100, IndexSize: 2048},
				dbStats{DB: "bar", Collections: 2, Objects: 30, AvgObjSize: 1030, DataSize: 1025, StorageSize: 1021, NumExtents: 11, Indexes: 101, IndexSize: 2049},
			},
			TestDescription: "Should successfully read two database's  stats from mongo",
		},
	}

	for _, test := range tests {
		g.Describe("readDBStats()", func() {
			g.It(test.TestDescription, func() {
				res := readDBStats(test.InputLog, test.InputSession)
				g.Assert(reflect.DeepEqual(res, test.ExpectedRes)).IsTrue()
			})
		})
	}
}

func TestReadServerStats(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputLog     *logrus.Logger
		InputSession Session
		ExpectedRes  struct {
			Host                        string
			MongoMetricsCursorOpenTotal int
		}
		TestDescription string
	}{
		{
			InputLog: logrus.New(),
			InputSession: NewMockSession(
				MockSessionResults{
					RunResult: []byte("{\"host\":\"g00000000652073\",\"version\":\"3.2.8\",\"process\":\"mongod\",\"pid\":7539,\"uptime\":20356,\"uptimeMillis\":20356433,\"uptimeEstimate\":20103,\"asserts\":{\"regular\":0,\"warning\":0,\"msg\":0,\"user\":4,\"rollovers\":0},\"connections\":{\"current\":1,\"available\":51199,\"totalCreated\":2188},\"extra_info\":{\"note\":\"fields vary by platform\",\"heap_usage_bytes\":61575592,\"page_faults\":0},\"globalLock\":{\"totalTime\":20356436000,\"currentQueue\":{\"total\":0,\"readers\":0,\"writers\":0},\"activeClients\":{\"total\":8,\"readers\":0,\"writers\":0}},\"locks\":{\"Global\":{\"acquireCount\":{\"r\":47790,\"w\":12,\"W\":4}},\"Database\":{\"acquireCount\":{\"r\":20690,\"R\":3197,\"W\":12}},\"Collection\":{\"acquireCount\":{\"r\":18897,\"w\":7}},\"Metadata\":{\"acquireCount\":{\"w\":1}}},\"network\":{\"bytesIn\":1306830,\"bytesOut\":20897949,\"numRequests\":14247},\"opcounters\":{\"insert\":4,\"query\":2131,\"update\":4,\"delete\":0,\"getmore\":0,\"command\":12122},\"opcountersRepl\":{\"insert\":0,\"query\":0,\"update\":0,\"delete\":0,\"getmore\":0,\"command\":0},\"storageEngine\":{\"name\":\"wiredTiger\",\"supportsCommittedReads\":true,\"persistent\":true},\"tcmalloc\":{\"generic\":{\"current_allocated_bytes\":61577128,\"heap_size\":66281472},\"tcmalloc\":{\"pageheap_free_bytes\":1933312,\"pageheap_unmapped_bytes\":0,\"max_total_thread_cache_bytes\":1073741824,\"current_total_thread_cache_bytes\":2107800,\"central_cache_free_bytes\":663232,\"transfer_cache_free_bytes\":0,\"thread_cache_free_bytes\":2107800,\"aggressive_memory_decommit\":0}},\"wiredTiger\":{\"LSM\":{\"application work units currently queued\":0,\"merge work units currently queued\":0,\"rows merged in an LSM tree\":0,\"sleep for LSM checkpoint throttle\":0,\"sleep for LSM merge throttle\":0,\"switch work units currently queued\":0,\"tree maintenance operations discarded\":0,\"tree maintenance operations executed\":0,\"tree maintenance operations scheduled\":0,\"tree queue hit maximum\":0},\"async\":{\"current work queue length\":0,\"maximum work queue length\":0,\"number of allocation state races\":0,\"number of flush calls\":0,\"number of operation slots viewed for allocation\":0,\"number of times operation allocation failed\":0,\"number of times worker found no work\":0,\"total allocations\":0,\"total compact calls\":0,\"total insert calls\":0,\"total remove calls\":0,\"total search calls\":0,\"total update calls\":0},\"block-manager\":{\"blocks pre-loaded\":0,\"blocks read\":1,\"blocks written\":34,\"bytes read\":4096,\"bytes written\":155648,\"mapped blocks read\":0,\"mapped bytes read\":0},\"cache\":{\"bytes currently in the cache\":37438,\"bytes read into cache\":0,\"bytes written from cache\":28989,\"checkpoint blocked page eviction\":0,\"eviction currently operating in aggressive mode\":0,\"eviction server candidate queue empty when topping up\":0,\"eviction server candidate queue not empty when topping up\":0,\"eviction server evicting pages\":0,\"eviction server populating queue, but not evicting pages\":0,\"eviction server unable to reach eviction goal\":0,\"eviction worker thread evicting pages\":0,\"failed eviction of pages that exceeded the in-memory maximum\":0,\"files with active eviction walks\":0,\"files with new eviction walks started\":0,\"hazard pointer blocked page eviction\":0,\"in-memory page passed criteria to be split\":0,\"in-memory page splits\":0,\"internal pages evicted\":0,\"internal pages split during eviction\":0,\"leaf pages split during eviction\":0,\"lookaside table insert calls\":0,\"lookaside table remove calls\":0,\"maximum bytes configured\":1073741824,\"maximum page size at eviction\":0,\"modified pages evicted\":0,\"modified pages evicted by application threads\":0,\"page split during eviction deepened the tree\":0,\"page written requiring lookaside records\":0,\"pages currently held in the cache\":21,\"pages evicted because they exceeded the in-memory maximum\":0,\"pages evicted because they had chains of deleted items\":0,\"pages evicted by application threads\":0,\"pages queued for eviction\":0,\"pages queued for urgent eviction\":0,\"pages read into cache\":0,\"pages read into cache requiring lookaside entries\":0,\"pages seen by eviction walk\":0,\"pages selected for eviction unable to be evicted\":0,\"pages walked for eviction\":0,\"pages written from cache\":22,\"pages written requiring in-memory restoration\":0,\"percentage overhead\":8,\"tracked bytes belonging to internal pages in the cache\":2723,\"tracked bytes belonging to leaf pages in the cache\":34715,\"tracked bytes belonging to overflow pages in the cache\":0,\"tracked dirty bytes in the cache\":0,\"tracked dirty pages in the cache\":0,\"unmodified pages evicted\":0},\"connection\":{\"auto adjusting condition resets\":16,\"auto adjusting condition wait calls\":61112,\"files currently open\":14,\"memory allocations\":539570,\"memory frees\":538659,\"memory re-allocations\":86347,\"pthread mutex condition wait calls\":267128,\"pthread mutex shared lock read-lock calls\":57797,\"pthread mutex shared lock write-lock calls\":26786,\"total read I/Os\":21,\"total write I/Os\":72},\"cursor\":{\"cursor create calls\":57,\"cursor insert calls\":69,\"cursor next calls\":15,\"cursor prev calls\":5,\"cursor remove calls\":0,\"cursor reset calls\":34782,\"cursor restarted searches\":0,\"cursor search calls\":34268,\"cursor search near calls\":464,\"cursor update calls\":0,\"truncate calls\":0},\"data-handle\":{\"connection data handles currently active\":11,\"connection sweep candidate became referenced\":0,\"connection sweep dhandles closed\":0,\"connection sweep dhandles removed from hash list\":3042,\"connection sweep time-of-death sets\":3042,\"connection sweeps\":2035,\"session dhandles swept\":0,\"session sweep attempts\":363},\"log\":{\"busy returns attempting to switch slots\":0,\"consolidated slot closures\":21,\"consolidated slot join races\":0,\"consolidated slot join transitions\":21,\"consolidated slot joins\":51,\"consolidated slot unbuffered writes\":0,\"log bytes of payload data\":15406,\"log bytes written\":19328,\"log files manually zero-filled\":0,\"log flush operations\":203273,\"log force write operations\":223987,\"log force write operations skipped\":223978,\"log records compressed\":19,\"log records not compressed\":16,\"log records too small to compress\":16,\"log release advances write LSN\":13,\"log scan operations\":0,\"log scan records requiring two reads\":0,\"log server thread advances write LSN\":8,\"log server thread write LSN walk skipped\":21351,\"log sync operations\":21,\"log sync_dir operations\":1,\"log write operations\":51,\"logging bytes consolidated\":18944,\"maximum log file size\":104857600,\"number of pre-allocated log files to create\":2,\"pre-allocated log files not ready and missed\":1,\"pre-allocated log files prepared\":2,\"pre-allocated log files used\":0,\"records processed by log scan\":0,\"total in-memory size of compressed records\":22950,\"total log buffer size\":33554432,\"total size of compressed records\":11353,\"written slots coalesced\":0,\"yields waiting for previous log file close\":0},\"reconciliation\":{\"fast-path pages deleted\":0,\"page reconciliation calls\":22,\"page reconciliation calls for eviction\":0,\"pages deleted\":0,\"split bytes currently awaiting free\":0,\"split objects currently awaiting free\":0},\"session\":{\"open cursor count\":28,\"open session count\":15},\"thread-yield\":{\"page acquire busy blocked\":0,\"page acquire eviction blocked\":0,\"page acquire locked blocked\":0,\"page acquire read blocked\":0,\"page acquire time sleeping (usecs)\":0},\"transaction\":{\"number of named snapshots created\":0,\"number of named snapshots dropped\":0,\"transaction begins\":6486,\"transaction checkpoint currently running\":0,\"transaction checkpoint generation\":339,\"transaction checkpoint max time (msecs)\":173,\"transaction checkpoint min time (msecs)\":0,\"transaction checkpoint most recent time (msecs)\":0,\"transaction checkpoint total time (msecs)\":184,\"transaction checkpoints\":339,\"transaction failures due to cache overflow\":0,\"transaction range of IDs currently pinned\":0,\"transaction range of IDs currently pinned by a checkpoint\":0,\"transaction range of IDs currently pinned by named snapshots\":0,\"transaction sync calls\":0,\"transactions committed\":9,\"transactions rolled back\":6477},\"concurrentTransactions\":{\"write\":{\"out\":0,\"available\":128,\"totalTickets\":128},\"read\":{\"out\":0,\"available\":128,\"totalTickets\":128}}},\"writeBacksQueued\":false,\"mem\":{\"bits\":64,\"resident\":38,\"virtual\":390,\"supported\":true,\"mapped\":0,\"mappedWithJournal\":0},\"metrics\":{\"cursor\":{\"timedOut\":0,\"open\":{\"noTimeout\":3,\"pinned\":2,\"total\":5}},\"document\":{\"deleted\":0,\"inserted\":4,\"returned\":0,\"updated\":4},\"getLastError\":{\"wtime\":{\"num\":0,\"totalMillis\":0},\"wtimeouts\":0},\"operation\":{\"fastmod\":0,\"idhack\":0,\"scanAndOrder\":0,\"writeConflicts\":0},\"queryExecutor\":{\"scanned\":3,\"scannedObjects\":3},\"record\":{\"moves\":0},\"repl\":{\"executor\":{\"counters\":{\"eventCreated\":0,\"eventWait\":0,\"cancels\":0,\"waits\":0,\"scheduledNetCmd\":0,\"scheduledDBWork\":0,\"scheduledXclWork\":0,\"scheduledWorkAt\":0,\"scheduledWork\":0,\"schedulingFailures\":0},\"queues\":{\"networkInProgress\":0,\"dbWorkInProgress\":0,\"exclusiveInProgress\":0,\"sleepers\":0,\"ready\":0,\"free\":0},\"unsignaledEvents\":0,\"eventWaiters\":0,\"shuttingDown\":false,\"networkInterface\":\"NetworkInterfaceASIO inShutdown: 0\"},\"apply\":{\"batches\":{\"num\":0,\"totalMillis\":0},\"ops\":0},\"buffer\":{\"count\":0,\"maxSizeBytes\":268435456,\"sizeBytes\":0},\"network\":{\"bytes\":0,\"getmores\":{\"num\":0,\"totalMillis\":0},\"ops\":0,\"readersCreated\":0},\"preload\":{\"docs\":{\"num\":0,\"totalMillis\":0},\"indexes\":{\"num\":0,\"totalMillis\":0}}},\"storage\":{\"freelist\":{\"search\":{\"bucketExhausted\":0,\"requests\":0,\"scanned\":0}}},\"ttl\":{\"deletedDocuments\":0,\"passes\":339}}}"),
				},
				map[string]MockDatabaseResults{},
			),
			ExpectedRes: struct {
				Host                        string
				MongoMetricsCursorOpenTotal int
			}{
				Host: "g00000000652073",
				MongoMetricsCursorOpenTotal: 5,
			},
		},
	}

	for _, test := range tests {
		g.Describe("readDBStats()", func() {
			g.It(test.TestDescription, func() {
				res := readServerStats(test.InputLog, test.InputSession)
				g.Assert(res.Host).Equal(test.ExpectedRes.Host)
				g.Assert(res.Metrics.Cursor.Open.Total).Equal(test.ExpectedRes.MongoMetricsCursorOpenTotal)
			})
		})
	}
}

func TestFatalIfErrt(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		InputLog        *logrus.Logger
		InputErr        error
		TestDescription string
	}{
		{
			InputLog:        logrus.New(),
			InputErr:        nil,
			TestDescription: "Should successfully not exit on a nil error",
		},
	}

	for _, test := range tests {
		g.Describe("fatalIfErr()", func() {
			g.It(test.TestDescription, func() {
				fatalIfErr(test.InputLog, test.InputErr)
				g.Assert(true).IsTrue()
			})
		})
	}
}
