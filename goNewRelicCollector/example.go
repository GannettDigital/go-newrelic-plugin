package collectors

import "github.com/GannettDigital/go-newrelic-plugin/collectors"

// ExampleCollector used for reference for collector developers
func ExampleCollector(config collectors.Config, stats chan<- map[string]interface{}) {
	// do something real to collect some stats for your specific technology //

	// Important:
	// If you error and are not able to gather data, you need to log and close the stats chan,
	// to signify that an error occured and no data should be sent
	// close(stats)

	// send your aggregated data back to the publisher
	stats <- map[string]interface{}{
		"example.derp.1": 50,
		"example.herp.2": 1,
	}
}
