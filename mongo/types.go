package mongo

//MongoConfig is the keeper of the config
type mongoConfig struct {
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
	commitsInWriteLock int `bson:"commitsInWriteLock"`
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
}
