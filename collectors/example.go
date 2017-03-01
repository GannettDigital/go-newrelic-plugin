package collectors

func ExampleCollector(config Config, stats chan<- map[string]interface{}) {
	// collect some stats //

	// Important:
	// If you error at all, you need to log and close the stats chan,
	// to signify that an error occured and no data should be sent
	// close(stats)

	// send your aggregated data back to the publisher
	stats <- map[string]interface{}{
		"example.derp.1": 50,
		"example.herp.2": 1,
	}
}
