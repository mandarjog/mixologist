package mixologist

import (
	sc "google/api/servicecontrol/v1"
)

type PrometheusReporter struct {
	ReportQueue chan sc.ReportRequest
}

func NewPrometheusReporter() ReportConsumer {
	return &PrometheusReporter{}
}

func (p *PrometheusReporter) SetReportQueue(ch chan sc.ReportRequest) {
	p.ReportQueue = ch
}

// Do prometheus specific processing here
// Start a new listener etc.
func (p *PrometheusReporter) Start() {

}

// Stop Processing
func (p *PrometheusReporter) Stop() {

}
