package mixologist

import (
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	sc "google/api/servicecontrol/v1"
)

type PrometheusReporter struct {
	ReportQueue chan sc.ReportRequest
	MetricMap   map[string]*prometheus.SummaryVec
}

func NewPrometheusReporter() ReportConsumer {
	latency := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "api_producer_total_latencies",
			Help: "serviceruntime.googleapis.com/api/producer/total_latencies",
		},
		[]string{"service"},
	)
	mm := make(map[string]*prometheus.SummaryVec)
	mm["serviceruntime.googleapis.com/api/producer/total_latencies"] = latency
	prometheus.MustRegister(latency)
	return &PrometheusReporter{
		MetricMap: mm,
	}
}

func (p *PrometheusReporter) SetReportQueue(ch chan sc.ReportRequest) {
	p.ReportQueue = ch
}

// Do prometheus specific processing here
// Start a new listener etc.
func (p *PrometheusReporter) Start() {
	glog.Info("Staring prometheus loop")
	for reportMsg := range p.ReportQueue {
		for _, oprn := range reportMsg.GetOperations() {
			for _, metric := range oprn.GetMetricValueSets() {
				if summvec, ok := p.MetricMap[metric.MetricName]; ok {
					val := metric.MetricValues[0].GetDoubleValue()
					glog.Info(val)
					glog.Info(metric.MetricValues[0])
					summvec.WithLabelValues("producer").Observe(val)
				}
			}
		}
	}
}

// Stop Processing
func (p *PrometheusReporter) Stop() {

}
