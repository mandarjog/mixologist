package prometheus

import (
	promclnt "github.com/prometheus/client_golang/prometheus"
	"somnacin-internal/mixologist/mixologist"
)

func init() {
	mixologist.RegisterReportConsumer(Name, new(builder))
}

type (
	prometheusConsumer struct {
		MetricSummaryMap map[string]*promclnt.SummaryVec
		ProducerMetrics  map[string]interface{}
	}
	builder struct {
	}
)
