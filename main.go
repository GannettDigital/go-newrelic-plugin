package main

import "time"

func main() {
	// list of collectors that exist
	// the key needs to match the value as defined in the config file
	// the value is the collector method that will be used to gather the stats for that type
	collectorArray := map[string]Collector{
		"nginx": nginxCollector,
	}

	// TODO: populate config
	config := Config{}

	// main routine
	for name, collector := range collectorArray {
		go func() {
			if _, exists := config[name]; exists {
				if config[name]["enabled"] == "true" || true {
					// TODO: random delay to offset collections
					// TODO: time sourced from config
					ticker := time.NewTicker(time.Millisecond * 500)
					for t := range ticker.C {
						go getResult(config, collector)
					}
				}
			}
		}()
	}

	done := make(chan bool)
	<-done // block forever

}

func getResult(config Config, collector Collector) {
	c := make(chan map[string]string, 1)
	collector(config, c)

	select {
	case res := <-c:
		sendData(config, res)
	case <-time.After(time.Second * 10):
		// timeout so we don't leaving threads that block forever
	}
}

func sendData(config Config, stats map[string]string) {
	// send stats
}
