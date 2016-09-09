package mixologist

import (
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	sc "google/api/servicecontrol/v1"
	"math/big"
	"net/http"
	"strings"
)

const (
	missingLabelValue = "UNKNOWN"
)

var (
	sizeHistogramBuckets = prometheus.ExponentialBuckets(1, 10, 8)
	timeHistogramBuckets = prometheus.ExponentialBuckets(1e-6, 10, 8)
	histogramLabelNames  = prometheusSafeNames(MonitoredAPIResourceLabels)
	slashReplacer        = strings.NewReplacer("/", "_", ".", "_")
)

type prometheusConsumer struct {
	ReportQueue      chan *sc.ReportRequest
	MetricSummaryMap map[string]*prometheus.SummaryVec
	ProducerMetrics  map[string]interface{}
}

func prometheusSafeName(n string) string {
	s := strings.TrimPrefix(n, "/")
	return slashReplacer.Replace(s)
}

func prometheusSafeNames(names []string) []string {
	out := make([]string, len(names), len(names))
	for i, v := range names {
		out[i] = prometheusSafeName(v)
	}
	return out
}

// prepare identity mapped s
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

func newCounterVec(name, desc string, labels ...string) *prometheus.CounterVec {
	c := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: prometheusSafeName(name),
			Help: desc,
		},
		append(histogramLabelNames, prometheusSafeNames(labels)...),
	)
	prometheus.MustRegister(c)
	return c
}

// {8, 10.0, 1e-6};
func newTimeHistogramVec(name, desc string) *prometheus.HistogramVec {
	c := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    prometheusSafeName(name),
			Help:    desc,
			Buckets: timeHistogramBuckets,
		},
		histogramLabelNames,
	)
	prometheus.MustRegister(c)
	return c
}

// {8, 10.0, 1};
func newSizeHistogramVec(name, desc string) *prometheus.HistogramVec {
	c := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    prometheusSafeName(name),
			Help:    desc,
			Buckets: sizeHistogramBuckets,
		},
		histogramLabelNames,
	)
	prometheus.MustRegister(c)
	return c
}

func producerMetrics() map[string]interface{} {
	m := map[string]interface{}{
		ProducerRequestCount:           newCounterVec(ProducerRequestCount, "Request Count", PerMetricLabels[ProducerRequestCount]...),
		ProducerTotalLatencies:         newTimeHistogramVec(ProducerTotalLatencies, "Total latencies"),
		ProducerBackendLatencies:       newTimeHistogramVec(ProducerBackendLatencies, "Backend latencies"),
		ProducerRequestSizes:           newSizeHistogramVec(ProducerRequestSizes, "Request Sizes"),
		ProducerRequestCountByConsumer: newCounterVec(ProducerRequestCountByConsumer, "Request Count By Consumer", PerMetricLabels[ProducerRequestCountByConsumer]...),
	}
	return m
}

// NewPrometheusConsumer -- obtain a new consumer that understands predefined mappings
func NewPrometheusConsumer() ReportConsumer {
	return &prometheusConsumer{
		MetricSummaryMap: getMetricsMap([][]string{
			[]string{"request_latency_in_ms", "http_request_duration_microseconds"},
			[]string{"request_size", "http_request_size_bytes"},
			[]string{"response_size", "http_response_size_bytes"}}),
		ProducerMetrics: producerMetrics(),
	}
}

func (p *prometheusConsumer) SetReportQueue(ch chan *sc.ReportRequest) {
	p.ReportQueue = ch
}

func buildLabels(defaults map[string]string, m *sc.MetricValue) prometheus.Labels {
	l := prometheus.Labels{}
	for _, k := range MonitoredAPIResourceLabels {
		l[prometheusSafeName(k)] = missingLabelValue
	}
	for k, v := range defaults {
		l[prometheusSafeName(k)] = v
	}
	for k, v := range m.GetLabels() {
		l[prometheusSafeName(k)] = v
	}
	return l
}

func bucketsMatch(b *sc.Distribution_ExponentialBuckets, d *ExponentialDistribution) bool {
	if b == nil {
		return false
	}
	return b.NumFiniteBuckets == d.NumBuckets &&
		big.NewFloat(b.GrowthFactor).Cmp(big.NewFloat(d.GrowthFactor)) == 0 &&
		big.NewFloat(b.Scale).Cmp(big.NewFloat(d.StartValue)) == 0
}

func populateFromBuckets(hist prometheus.Histogram, dist *sc.Distribution, ed *ExponentialDistribution, buckets []float64) {
	if !bucketsMatch(dist.GetExponentialBuckets(), ed) {
		glog.Warningf("not a match. expected: %v, got: %\n", ed, dist.GetExponentialBuckets)
		return
	}

	curr := dist.Minimum
	if curr == 0 {
		curr = buckets[0]
	}
	for i, v := range dist.BucketCounts {
		if i > 0 && i < len(buckets) {
			curr = buckets[i-1] + (buckets[i]-buckets[i-1])/2
		}
		if i >= len(buckets) {
			if curr = dist.Maximum; curr == 0 {
				curr = buckets[len(buckets)-1] * ed.GrowthFactor / 2
			}

		}
		for k := 0; int64(k) < v; k++ {
			hist.Observe(curr)
		}
	}
}

func filter(src prometheus.Labels, metricLabels []string) prometheus.Labels {
	filteredLabels := prometheus.Labels{}
	for _, k := range metricLabels {
		if v, ok := src[k]; ok {
			filteredLabels[k] = v
		}
	}
	return filteredLabels
}

func addObservation(h *prometheus.HistogramVec, l prometheus.Labels, mv *sc.MetricValue, name string) {
	d := mv.GetDistributionValue()
	if d == nil {
		return
	}
	switch name {
	case ProducerTotalLatencies, ProducerBackendLatencies:
		// time buckets
		populateFromBuckets(h.With(filter(l, histogramLabelNames)), d, TimeDistribution, timeHistogramBuckets)
	case ProducerRequestSizes:
		// size buckets
		populateFromBuckets(h.With(filter(l, histogramLabelNames)), d, SizeDistribution, sizeHistogramBuckets)
	default:
		glog.Warningf("unknown metric for distribution: %s", name)
	}
}

func (p *prometheusConsumer) process(mvs *sc.MetricValueSet, defaultLabels map[string]string) {
	if m, ok := p.ProducerMetrics[mvs.MetricName]; ok {
		for _, mv := range mvs.GetMetricValues() {
			labels := buildLabels(defaultLabels, mv)
			switch t := m.(type) {
			case *prometheus.CounterVec:

				// need to ensure proper label cardinality
				filteredLabels := filter(labels, histogramLabelNames)
				for _, k := range PerMetricLabels[mvs.MetricName] {
					if v, ok := labels[prometheusSafeName(k)]; ok {
						filteredLabels[prometheusSafeName(k)] = v
					} else {
						filteredLabels[prometheusSafeName(k)] = missingLabelValue
					}
				}

				t.With(filteredLabels).Add(float64(mv.GetInt64Value()))
			case *prometheus.HistogramVec:
				addObservation(t, labels, mv, mvs.MetricName)
			}
		}
	}
}

// This loop will not exit until the channel closes
func (p *prometheusConsumer) consumerLoop() {
	glog.Info("Starting prometheus loop")
	for reportMsg := range p.ReportQueue {
		service := reportMsg.ServiceName
		for _, oprn := range reportMsg.GetOperations() {
			defaultLabels := oprn.GetLabels()
			defaultLabels[CloudService] = service
			defaultLabels[ConsumerID] = oprn.ConsumerId

			for _, mvs := range oprn.GetMetricValueSets() {
				p.process(mvs, defaultLabels)
			}

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
