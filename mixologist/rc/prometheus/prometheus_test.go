package prometheus

import (
	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	sc "google/api/servicecontrol/v1"
	"somnacin-internal/mixologist/mixologist"
	"testing"
)

const (
	testSvc        = "testtestSvc"
	testLoc        = "us-east1-b"
	testAPIVersion = testSvc + "-v1.appspot.com"
	testAPIMethod  = "ListShelves"
)

var (
	testOpLbls = map[string]string{
		mixologist.CloudService:  testSvc,
		mixologist.CloudLocation: testLoc,
		mixologist.APIVersion:    testAPIVersion,
		mixologist.APIMethod:     testAPIMethod,
	}
	testPromLbls = prometheus.Labels{
		prometheusSafeName(mixologist.CloudService):      testSvc,
		prometheusSafeName(mixologist.CloudLocation):     testLoc,
		prometheusSafeName(mixologist.APIVersion):        testAPIVersion,
		prometheusSafeName(mixologist.APIMethod):         testAPIMethod,
		prometheusSafeName(mixologist.Protocol):          missingLabelValue,
		prometheusSafeName(mixologist.ResponseCode):      missingLabelValue,
		prometheusSafeName(mixologist.ResponseCodeClass): missingLabelValue,
		prometheusSafeName(mixologist.StatusCode):        missingLabelValue,
		prometheusSafeName(mixologist.CloudUid):          missingLabelValue,
		prometheusSafeName(mixologist.ConsumerProject):   missingLabelValue,
		prometheusSafeName(mixologist.CloudProject):      missingLabelValue,
	}
	testDistVals = []int64{1, 0, 1, 0, 0, 3, 0, 2, 0, 1}
	testDistSum  = 6.1165056e+07
)

func TestConsume(t *testing.T) {
	var consumeTests = []struct {
		name    string
		report  *sc.ReportRequest
		metrics map[string]*dto.Metric
	}{
		// TODO(dougreid): add more tests (ignored metrics, etc.)
		{
			name:    "Requests",
			report:  reportReq(testSvc, operation(testOpLbls, metricValueSet(mixologist.ProducerRequestCount, metricValue(14)))),
			metrics: map[string]*dto.Metric{mixologist.ProducerRequestCount: newCounterMetricProto(14)},
		},
		{
			name:    "Requests -- Multiple MetricValues",
			report:  reportReq(testSvc, operation(testOpLbls, metricValueSet(mixologist.ProducerRequestCount, metricValue(14), metricValue(22)))),
			metrics: map[string]*dto.Metric{mixologist.ProducerRequestCount: newCounterMetricProto(36)},
		},
		{
			name:    "Total Latencies",
			report:  reportReq(testSvc, operation(testOpLbls, metricValueSet(mixologist.ProducerTotalLatencies, timeDistMetricValue(testDistVals)))),
			metrics: map[string]*dto.Metric{mixologist.ProducerTotalLatencies: newHistogramMetricProto(testDistVals, testDistSum*mixologist.TimeDistributionScale, timeHistogramBuckets)},
		},
		{
			name:    "Total Latencies -- Multiple MetricValues",
			report:  reportReq(testSvc, operation(testOpLbls, metricValueSet(mixologist.ProducerTotalLatencies, timeDistMetricValue(testDistVals)), metricValueSet(mixologist.ProducerTotalLatencies, timeDistMetricValue(testDistVals)))),
			metrics: map[string]*dto.Metric{mixologist.ProducerTotalLatencies: newHistogramMetricProto([]int64{2, 0, 2, 0, 0, 6, 0, 4, 0, 2}, 2*testDistSum*mixologist.TimeDistributionScale, timeHistogramBuckets)},
		},
		{
			name:    "Request Sizes",
			report:  reportReq(testSvc, operation(testOpLbls, metricValueSet(mixologist.ProducerRequestSizes, sizeDistMetricValue(testDistVals)))),
			metrics: map[string]*dto.Metric{mixologist.ProducerRequestSizes: newHistogramMetricProto(testDistVals, testDistSum*mixologist.SizeDistributionScale, sizeHistogramBuckets)},
		},
		{
			name:   "Mixed Metrics",
			report: reportReq(testSvc, operation(testOpLbls, metricValueSet(mixologist.ProducerRequestCount, metricValue(347)), metricValueSet(mixologist.ProducerRequestSizes, sizeDistMetricValue(testDistVals)))),
			metrics: map[string]*dto.Metric{
				mixologist.ProducerRequestCount: newCounterMetricProto(347),
				mixologist.ProducerRequestSizes: newHistogramMetricProto(testDistVals, testDistSum*mixologist.SizeDistributionScale, sizeHistogramBuckets),
			},
		},
	}
	p := &consumer{}
	for _, v := range consumeTests {
		p.Consume(v.report)
		for name, want := range v.metrics {
			got := p.labelStrippedMetric(name)
			if !proto.Equal(got, want) {
				t.Errorf("%s: metrics not equal; got %v, want %v", v.name, got, want)
			}
		}
		// clean-up / deregister to prevent metric collisions
		for _, met := range metrics {
			if r, ok := met.(resetter); ok {
				r.Reset()
			}
		}
	}
}

// prometheus.MetricVec is a struct, not an interface
// create a simple interface here to allow easy resetting
type resetter interface {
	Reset()
}

func (p *consumer) labelStrippedMetric(name string) *dto.Metric {
	m := &dto.Metric{}
	switch name {
	case mixologist.ProducerRequestCount:
		metrics[name].(*prometheus.CounterVec).With(testPromLbls).Write(m)
	case mixologist.ProducerTotalLatencies, mixologist.ProducerRequestSizes:
		l := prometheus.Labels{}
		for _, k := range histogramLabelNames {
			l[k] = testPromLbls[k]
		}
		metrics[name].(*prometheus.HistogramVec).With(l).Write(m)
	}
	m.Label = []*dto.LabelPair{}
	return m
}

func metricValue(v int64) *sc.MetricValue {
	return &sc.MetricValue{Value: &sc.MetricValue_Int64Value{Int64Value: v}}
}

func timeDistMetricValue(bucketCounts []int64) *sc.MetricValue {
	return &sc.MetricValue{
		Value: &sc.MetricValue_DistributionValue{
			DistributionValue: &sc.Distribution{
				Count:        8,
				BucketCounts: bucketCounts,
				BucketOption: &sc.Distribution_ExponentialBuckets_{
					ExponentialBuckets: &sc.Distribution_ExponentialBuckets{
						NumFiniteBuckets: mixologist.TimeDistributionBuckets,
						GrowthFactor:     mixologist.TimeDistributionGrowthFactor,
						Scale:            mixologist.TimeDistributionScale,
					},
				},
			},
		},
	}
}

func sizeDistMetricValue(bucketCounts []int64) *sc.MetricValue {
	return &sc.MetricValue{
		Value: &sc.MetricValue_DistributionValue{
			DistributionValue: &sc.Distribution{
				Count:        8,
				BucketCounts: bucketCounts,
				BucketOption: &sc.Distribution_ExponentialBuckets_{
					ExponentialBuckets: &sc.Distribution_ExponentialBuckets{
						NumFiniteBuckets: mixologist.SizeDistributionBuckets,
						GrowthFactor:     mixologist.SizeDistributionGrowthFactor,
						Scale:            mixologist.SizeDistributionScale,
					},
				},
			},
		},
	}
}

func metricValueSet(name string, mv ...*sc.MetricValue) *sc.MetricValueSet {
	return &sc.MetricValueSet{MetricName: name, MetricValues: mv}
}

func operation(labels map[string]string, mvs ...*sc.MetricValueSet) *sc.Operation {
	return &sc.Operation{Labels: labels, MetricValueSets: mvs}
}

func reportReq(name string, ops ...*sc.Operation) *sc.ReportRequest {
	return &sc.ReportRequest{ServiceName: name, Operations: ops}
}

func newCounterProto(v float64) *dto.Counter {
	return &dto.Counter{Value: proto.Float64(v)}
}

func newCounterMetricProto(v float64) *dto.Metric {
	return &dto.Metric{Counter: newCounterProto(v)}
}

func newBuckets(counts []int64, bounds []float64) []*dto.Bucket {
	b := []*dto.Bucket{}
	var c int64
	for i, v := range bounds {
		c = c + counts[i]
		b = append(b, &dto.Bucket{CumulativeCount: proto.Uint64(uint64(c)), UpperBound: proto.Float64(v)})
	}
	return b
}

func newHistogramProto(count int64, sum float64, b []*dto.Bucket) *dto.Histogram {
	return &dto.Histogram{SampleCount: proto.Uint64(uint64(count)), SampleSum: proto.Float64(sum), Bucket: b}
}

func newHistogramMetricProto(counts []int64, sum float64, bounds []float64) *dto.Metric {
	var count int64
	for _, v := range counts {
		count += v
	}
	return &dto.Metric{Histogram: newHistogramProto(count, sum, newBuckets(counts, bounds))}
}
