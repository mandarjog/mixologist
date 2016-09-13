package prometheus

import (
	"github.com/golang/glog"
	promclnt "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	sc "google/api/servicecontrol/v1"
	"math/big"
	"somnacin-internal/mixologist/mixologist"
	"strings"
)

var (
	sizeHistogramBuckets = promclnt.ExponentialBuckets(1, 10, 8)
	timeHistogramBuckets = promclnt.ExponentialBuckets(1e-6, 10, 8)
	histogramLabelNames  = prometheusSafeNames(mixologist.MonitoredAPIResourceLabels)
	slashReplacer        = strings.NewReplacer("/", "_", ".", "_")
)

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
func getMetricsMap(metricNames [][]string) map[string]*promclnt.SummaryVec {
	mm := make(map[string]*promclnt.SummaryVec)
	for _, metricName := range metricNames {
		m := promclnt.NewSummaryVec(
			promclnt.SummaryOpts{
				Name: metricName[1],
				Help: metricName[0],
			},
			[]string{"service", "api_method"},
		)
		mm[metricName[0]] = m
		promclnt.MustRegister(m)
	}
	return mm
}

func newCounterVec(name, desc string, labels ...string) *promclnt.CounterVec {
	c := promclnt.NewCounterVec(
		promclnt.CounterOpts{
			Name: prometheusSafeName(name),
			Help: desc,
		},
		append(histogramLabelNames, prometheusSafeNames(labels)...),
	)
	promclnt.MustRegister(c)
	return c
}

// {8, 10.0, 1e-6};
func newTimeHistogramVec(name, desc string) *promclnt.HistogramVec {
	c := promclnt.NewHistogramVec(
		promclnt.HistogramOpts{
			Name:    prometheusSafeName(name),
			Help:    desc,
			Buckets: timeHistogramBuckets,
		},
		histogramLabelNames,
	)
	promclnt.MustRegister(c)
	return c
}

// {8, 10.0, 1};
func newSizeHistogramVec(name, desc string) *promclnt.HistogramVec {
	c := promclnt.NewHistogramVec(
		promclnt.HistogramOpts{
			Name:    prometheusSafeName(name),
			Help:    desc,
			Buckets: sizeHistogramBuckets,
		},
		histogramLabelNames,
	)
	promclnt.MustRegister(c)
	return c
}

func producerMetrics() map[string]interface{} {
	m := map[string]interface{}{
		mixologist.ProducerRequestCount:           newCounterVec(mixologist.ProducerRequestCount, "Request Count", mixologist.PerMetricLabels[mixologist.ProducerRequestCount]...),
		mixologist.ProducerTotalLatencies:         newTimeHistogramVec(mixologist.ProducerTotalLatencies, "Total latencies"),
		mixologist.ProducerBackendLatencies:       newTimeHistogramVec(mixologist.ProducerBackendLatencies, "Backend latencies"),
		mixologist.ProducerRequestSizes:           newSizeHistogramVec(mixologist.ProducerRequestSizes, "Request Sizes"),
		mixologist.ProducerRequestCountByConsumer: newCounterVec(mixologist.ProducerRequestCountByConsumer, "Request Count By Consumer", mixologist.PerMetricLabels[mixologist.ProducerRequestCountByConsumer]...),
	}
	return m
}

func buildLabels(defaults map[string]string, m *sc.MetricValue) promclnt.Labels {
	l := promclnt.Labels{}
	for _, k := range mixologist.MonitoredAPIResourceLabels {
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

func bucketsMatch(b *sc.Distribution_ExponentialBuckets, d *mixologist.ExponentialDistribution) bool {
	if b == nil {
		return false
	}
	return b.NumFiniteBuckets == d.NumBuckets &&
		big.NewFloat(b.GrowthFactor).Cmp(big.NewFloat(d.GrowthFactor)) == 0 &&
		big.NewFloat(b.Scale).Cmp(big.NewFloat(d.StartValue)) == 0
}

func populateFromBuckets(hist promclnt.Histogram, dist *sc.Distribution, ed *mixologist.ExponentialDistribution, buckets []float64) {
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

func filter(src promclnt.Labels, metricLabels []string) promclnt.Labels {
	filteredLabels := promclnt.Labels{}
	for _, k := range metricLabels {
		if v, ok := src[k]; ok {
			filteredLabels[k] = v
		}
	}
	return filteredLabels
}

func addObservation(h *promclnt.HistogramVec, l promclnt.Labels, mv *sc.MetricValue, name string) {
	d := mv.GetDistributionValue()
	if d == nil {
		return
	}
	switch name {
	case mixologist.ProducerTotalLatencies, mixologist.ProducerBackendLatencies:
		// time buckets
		populateFromBuckets(h.With(filter(l, histogramLabelNames)), d, mixologist.TimeDistribution, timeHistogramBuckets)
	case mixologist.ProducerRequestSizes:
		// size buckets
		populateFromBuckets(h.With(filter(l, histogramLabelNames)), d, mixologist.SizeDistribution, sizeHistogramBuckets)
	default:
		glog.Warningf("unknown metric for distribution: %s", name)
	}
}

func (p *prometheusConsumer) process(mvs *sc.MetricValueSet, defaultLabels map[string]string) {
	if m, ok := p.ProducerMetrics[mvs.MetricName]; ok {
		for _, mv := range mvs.GetMetricValues() {
			labels := buildLabels(defaultLabels, mv)
			switch t := m.(type) {
			case *promclnt.CounterVec:

				// need to ensure proper label cardinality
				filteredLabels := filter(labels, histogramLabelNames)
				for _, k := range mixologist.PerMetricLabels[mvs.MetricName] {
					if v, ok := labels[prometheusSafeName(k)]; ok {
						filteredLabels[prometheusSafeName(k)] = v
					} else {
						filteredLabels[prometheusSafeName(k)] = missingLabelValue
					}
				}

				t.With(filteredLabels).Add(float64(mv.GetInt64Value()))
			case *promclnt.HistogramVec:
				addObservation(t, labels, mv, mvs.MetricName)
			}
		}
	}
}

// Consume -- Called to consume 1 reportMsg at a time
func (p *prometheusConsumer) Consume(reportMsg *sc.ReportRequest) error {
	service := reportMsg.ServiceName
	for _, oprn := range reportMsg.GetOperations() {
		defaultLabels := oprn.GetLabels()
		defaultLabels[mixologist.CloudService] = service
		defaultLabels[mixologist.ConsumerID] = oprn.ConsumerId

		for _, mvs := range oprn.GetMetricValueSets() {
			p.process(mvs, defaultLabels)
		}

		for _, le := range oprn.GetLogEntries() {
			fm := le.GetStructPayload().GetFields()
			labels := promclnt.Labels{"api_method": fm["api_method"].GetStringValue(),
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

	//TODO return error when it makes sense
	return nil
}

// GetName interface method
func (p *prometheusConsumer) GetName() string {
	return Name
}

// Start a new listener etc.
// Push type (statsd) consumers will only have a consumer loop
// Prometheus needs an additional listerner so that the prometheus
// framework can fetch data from the /metrics endpoint
func (p *prometheusConsumer) GetPrefixAndHandler() *mixologist.PrefixAndHandler {
	return &mixologist.PrefixAndHandler{
		Prefix:  "/metrics",
		Handler: promhttp.Handler(),
	}
}

/* Implements ReportConsumerBuilder */
// New -- Returns a new prometheus consumer
func (s *builder) NewConsumer(meta map[string]interface{}) mixologist.ReportConsumer {
	return &prometheusConsumer{
		MetricSummaryMap: getMetricsMap([][]string{
			[]string{"request_latency_in_ms", "http_request_duration_microseconds"},
			[]string{"request_size", "http_request_size_bytes"},
			[]string{"response_size", "http_response_size_bytes"}}),
		ProducerMetrics: producerMetrics(),
		meta:            meta,
	}
}
