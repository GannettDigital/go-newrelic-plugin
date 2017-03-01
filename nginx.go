package goNewRelicCollector

// TODO:
func nginxCollector(config Config, stats chan<- map[string]string) {
	/*
	  When called, this needs to:
	    1. collect metrics from nginx
	    2. format metris into a map[string]string
	    3. send that map[string]string back on the channel (where the dispatcher will push it to NR)
	    4. thats all!
	*/
}
