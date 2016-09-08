package mixologist

import (
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	sc "google/api/servicecontrol/v1"
	"net/http"
)

type prometheusConsumer struct {
	ReportQueue      chan *sc.ReportRequest
	MetricSummaryMap map[string]*prometheus.SummaryVec
}

// prepare identity mapped metrics
func getMetricsMap(metricNames [][]string) map[string]*prometheus.SummaryVec {
	mm := make(map[string]*prometheus.SummaryVec)
	for _, metricName := range metricNames {
		m := prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: metricName[1],
				Help: metricName[0],
			},
			[]string{"service", "api_method"},
		)
		mm[metricName[0]] = m
		prometheus.MustRegister(m)
	}
	return mm
}

// NewPrometheusConsumer -- obtain a new consumer that understands predefined mappings
func NewPrometheusConsumer() ReportConsumer {
	return &prometheusConsumer{
		MetricSummaryMap: getMetricsMap([][]string{
			[]string{"request_latency_in_ms", "http_request_duration_microseconds"},
			[]string{"request_size", "http_request_size_bytes"},
			[]string{"response_size", "http_response_size_bytes"}}),
	}
}

func (p *prometheusConsumer) SetReportQueue(ch chan *sc.ReportRequest) {
	p.ReportQueue = ch
}

// This loop will not exit until the channel closes
func (p *prometheusConsumer) consumerLoop() {
	glog.Info("Starting prometheus loop")
	for reportMsg := range p.ReportQueue {
		for _, oprn := range reportMsg.GetOperations() {
			for _, le := range oprn.GetLogEntries() {
				fm := le.GetStructPayload().GetFields()
				labels := prometheus.Labels{"api_method": fm["api_method"].GetStringValue(),
					"service": fm["api_name"].GetStringValue()}
				for mn, metric := range p.MetricSummaryMap {
					if v := fm[mn]; v != nil {
						metric.With(labels).Observe(v.GetNumberValue())
					}
				}
			}
			if glog.V(1) {
				glog.Info("Processed ", len(oprn.GetLogEntries()), " Entries")
			}
		}
	}
}

// Do prometheus specific processing here
func (p *prometheusConsumer) Start() {
	p.consumerLoop()
}

// Start a new listener etc.
// Push type (statsd) consumers will only have a consumer loop
// Prometheus needs an additional listerner so that the prometheus
// framework can fetch data from the /metrics endpoint
// For Consumers that
func (p *prometheusConsumer) GetPrefixAndHandler() (string, http.Handler) {
	return "/metrics", promhttp.Handler()
}

// Stop Processing
func (p *prometheusConsumer) Stop() {

}
