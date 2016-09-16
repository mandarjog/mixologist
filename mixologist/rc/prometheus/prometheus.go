package prometheus

import (
	"github.com/golang/glog"
	pc "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	sc "google/api/servicecontrol/v1"
	"math/big"
	"somnacin-internal/mixologist/mixologist"
	"strings"
)

var (
	sizeHistogramBuckets = pc.ExponentialBuckets(1, 10, 8)
	timeHistogramBuckets = pc.ExponentialBuckets(1e-6, 10, 8)
	histogramLabelNames  = prometheusSafeNames(mixologist.MonitoredAPIResourceLabels)
	slashReplacer        = strings.NewReplacer("/", "_", ".", "_")

	// metrics
	reqs       = newCounterVec(mixologist.ProducerRequestCount, "Request Count", mixologist.PerMetricLabels[mixologist.ProducerRequestCount])
	reqsByCons = newCounterVec(mixologist.ProducerRequestCountByConsumer, "Request Count By Consumer", mixologist.PerMetricLabels[mixologist.ProducerRequestCountByConsumer])
	totLat     = newTimeHistogramVec(mixologist.ProducerTotalLatencies, "Total Latency")
	backendLat = newTimeHistogramVec(mixologist.ProducerBackendLatencies, "Backend Latency")
	reqSize    = newSizeHistogramVec(mixologist.ProducerRequestSizes, "Request Size")

	metrics = map[string]interface{}{
		mixologist.ProducerRequestCount:           reqs,
		mixologist.ProducerRequestCountByConsumer: reqsByCons,
		mixologist.ProducerTotalLatencies:         totLat,
		mixologist.ProducerBackendLatencies:       backendLat,
		mixologist.ProducerRequestSizes:           reqSize,
	}
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
func getMetricsMap(metricNames [][]string) map[string]*pc.SummaryVec {
	mm := make(map[string]*pc.SummaryVec)
	for _, metricName := range metricNames {
		m := pc.NewSummaryVec(
			pc.SummaryOpts{
				Name: metricName[1],
				Help: metricName[0],
			},
			[]string{"service", "api_method"},
		)
		mm[metricName[0]] = m
		pc.MustRegister(m)
	}
	return mm
}

func newCounterVec(name, desc string, labels []string) *pc.CounterVec {
	c := pc.NewCounterVec(
		pc.CounterOpts{
			Name: prometheusSafeName(name),
			Help: desc,
		},
		append(histogramLabelNames, prometheusSafeNames(labels)...),
	)
	return c
}

// {8, 10.0, 1e-6};
func newTimeHistogramVec(name, desc string) *pc.HistogramVec {
	c := pc.NewHistogramVec(
		pc.HistogramOpts{
			Name:    prometheusSafeName(name),
			Help:    desc,
			Buckets: timeHistogramBuckets,
		},
		histogramLabelNames,
	)
	return c
}

// {8, 10.0, 1};
func newSizeHistogramVec(name, desc string) *pc.HistogramVec {
	c := pc.NewHistogramVec(
		pc.HistogramOpts{
			Name:    prometheusSafeName(name),
			Help:    desc,
			Buckets: sizeHistogramBuckets,
		},
		histogramLabelNames,
	)
	return c
}

func buildLabels(defaults map[string]string, m *sc.MetricValue) pc.Labels {
	l := pc.Labels{}
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

func populateFromBuckets(hist pc.Histogram, dist *sc.Distribution, ed *mixologist.ExponentialDistribution, buckets []float64) {
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

func filter(src pc.Labels, metricLabels []string) pc.Labels {
	filteredLabels := pc.Labels{}
	for _, k := range metricLabels {
		if v, ok := src[k]; ok {
			filteredLabels[k] = v
		}
	}
	return filteredLabels
}

func addObservation(h *pc.HistogramVec, l pc.Labels, mv *sc.MetricValue, name string) {
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

func process(mvs *sc.MetricValueSet, defaultLabels map[string]string) {
	if m, ok := metrics[mvs.MetricName]; ok {
		for _, mv := range mvs.GetMetricValues() {
			labels := buildLabels(defaultLabels, mv)
			switch t := m.(type) {
			case *pc.CounterVec:

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
			case *pc.HistogramVec:
				addObservation(t, labels, mv, mvs.MetricName)
			}
		}
	}
}

// Consume -- Called to consume 1 reportMsg at a time
func (p *consumer) Consume(reportMsg *sc.ReportRequest) error {
	service := reportMsg.ServiceName
	for _, oprn := range reportMsg.GetOperations() {
		defaultLabels := oprn.GetLabels()
		defaultLabels[mixologist.CloudService] = service
		defaultLabels[mixologist.ConsumerID] = oprn.ConsumerId

		for _, mvs := range oprn.GetMetricValueSets() {
			process(mvs, defaultLabels)
		}

		for _, le := range oprn.GetLogEntries() {
			fm := le.GetStructPayload().GetFields()
			labels := pc.Labels{"api_method": fm["api_method"].GetStringValue(),
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
func (p *consumer) GetName() string {
	return Name
}

// Start a new listener etc.
// Push type (statsd) consumers will only have a consumer loop
// Prometheus needs an additional listerner so that the prometheus
// framework can fetch data from the /metrics endpoint
func (p *consumer) GetPrefixAndHandler() *mixologist.PrefixAndHandler {
	return &mixologist.PrefixAndHandler{
		Prefix:  "/metrics",
		Handler: promhttp.Handler(),
	}
}

/* Implements ReportConsumerBuilder */
// New -- Returns a new prometheus consumer
func (s *builder) NewConsumer(c mixologist.Config) mixologist.ReportConsumer {
	// only register when actually built
	for _, m := range metrics {
		pc.MustRegister(m.(pc.Collector))
	}
	return &consumer{
		MetricSummaryMap: getMetricsMap([][]string{
			[]string{"request_latency_in_ms", "http_request_duration_microseconds"},
			[]string{"request_size", "http_request_size_bytes"},
			[]string{"response_size", "http_response_size_bytes"}}),
	}
}
