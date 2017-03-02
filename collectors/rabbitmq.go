package collectors

import (
	"github.com/Sirupsen/logrus"
	RabbitHole "github.com/michaelklishin/rabbit-hole"
)

func getRabbitmqStatus(config RabbitmqConfig) (map[string]interface{}, error) {
	//statusUri := fmt.Sprintf("%v:%v/%v", config.Host, config.Port)
	//httpReq, err := http.NewRequest("GET", statusUri, bytes.NewBuffer([]byte("")))
	// URI, username, password
	rmqc, _ := RabbitHole.NewClient("http://10.84.100.176:15672", "scalr", "hiTVPamzPm")
	//	httpReq, err := http.NewRequest("GET", , bytes.NewBuffer([]byte("")))
	res, err := rmqc.Overview()
	// http.NewRequest error
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Encountered error creating http.NewRequest")

		return make(map[string]interface{}), err
	}
	log.WithFields(logrus.Fields{
		"error": res,
	}).Info("rabbit Over View")

	return map[string]interface{}{
		"rabbitmq.node.fd_used":      res.ErlangVersion,
		"rabbitmq.node.mem_used":     res.MessageStats,
		"rabbitmq.node.run_queue":    res.QueueTotals,
		"rabbitmq.node.sockets_used": res.MessageStats,
	}, nil
}

//RabbitmqCollector gets the rabbits stats.
func RabbitmqCollector(config Config, stats chan<- map[string]interface{}) {
	rabbitResponses, getStatsError := getRabbitmqStatus(config.RabbitmqConfig)
	if getStatsError != nil {
		close(stats)
		return
	}
	stats <- rabbitResponses
	/*
	  When called, this needs to:
	    1. collect metrics from rabbit
	    2. format metris into a map[string]interface{}
	    3. send that map[string]interface{} back on the channel (where the dispatcher will push it to NR)
	    4. thats all!
	*/
}
