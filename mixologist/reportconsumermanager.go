package mixologist

import (
	"github.com/golang/glog"
	sc "google/api/servicecontrol/v1"
)

// NewReportConsumerManager -- create a new report consumer manager with the configured list of consumers
func NewReportConsumerManager(rq chan *sc.ReportRequest, registry map[string]ReportConsumerBuilder, c Config) *ReportConsumerManagerImpl {
	glog.Infof("creating consumer manager with config: %v", c)
	consumerImpls := make([]ReportConsumer, 0, len(c.Metrics.Backends))
	for _, consumerName := range c.Metrics.Backends {
		if cn, ok := registry[consumerName]; ok {
			//TODO pass map params to cn.New
			consumerImpls = append(consumerImpls, cn.NewConsumer(c))
		}
	}
	return &ReportConsumerManagerImpl{
		reportQueue: rq,
		consumers:   consumerImpls,
	}
}

// Start -- consumer loop. Start the specified number of threads of consumer manager
func (s *ReportConsumerManagerImpl) Start(nConsumers int) {
	glog.Infof("Starting %d ConsumerLoops", nConsumers)
	for i := 0; i < nConsumers; i++ {
		go s.consumerLoop()
	}
}

// ConsumerLoop -- Start consumer loop. This method does not exit
func (s *ReportConsumerManagerImpl) consumerLoop() {
	for reportMsg := range s.reportQueue {
		for _, cc := range s.consumers {
			cc.Consume(reportMsg)
		}
	}
}

// GetPrefixAndHandlers -- Gather all the prefixes and handler from consumer, if any
func (s *ReportConsumerManagerImpl) GetPrefixAndHandlers() []*PrefixAndHandler {
	retval := make([]*PrefixAndHandler, 0, len(s.consumers))
	for _, cc := range s.consumers {
		if ph := cc.GetPrefixAndHandler(); ph != nil {
			retval = append(retval, ph)
		}
	}
	return retval
}
