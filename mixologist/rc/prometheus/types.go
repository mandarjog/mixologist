package prometheus

import (
	pc "github.com/prometheus/client_golang/prometheus"
	"github.com/cloudendpoints/mixologist/mixologist"
)

func init() {
	mixologist.RegisterReportConsumer(Name, new(builder))
}

type (
	consumer struct {
		MetricSummaryMap map[string]*pc.SummaryVec
	}
	builder struct {
	}
)
