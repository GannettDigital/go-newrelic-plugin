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

type serverStatus struct {
	Name string `json:"name"`
}
