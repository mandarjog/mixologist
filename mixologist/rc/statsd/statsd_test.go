package statsd

import (
	sd "github.com/cactus/go-statsd-client/statsd"
	sc "google/api/servicecontrol/v1"
	"reflect"
	"somnacin-internal/mixologist/mixologist"
	"testing"
)

const (
	svc = "test-api-service"
)

var (
	shelfLbls = map[string]string{
		mixologist.CloudService:  svc,
		mixologist.CloudLocation: "us-east1",
		mixologist.APIVersion:    "test-api-service-v1.appspot.com",
		mixologist.APIMethod:     "ListShelves",
	}
	bookLbls = map[string]string{
		mixologist.CloudService:  svc,
		mixologist.CloudLocation: "us-east1",
		mixologist.APIVersion:    "test-api-service-v1.appspot.com",
		mixologist.APIMethod:     "CreateBook",
	}
)

type fakeStatter struct {
	sd.Statter

	metrics map[string][]int64
}

func (f *fakeStatter) Inc(m string, v int64, s float32) error {
	f.metrics[m] = append(f.metrics[m], v)
	return nil
}

func (f *fakeStatter) Timing(m string, v int64, s float32) error {
	f.metrics[m] = append(f.metrics[m], v)
	return nil
}

func metricValue(v int64) *sc.MetricValue {
	return &sc.MetricValue{Value: &sc.MetricValue_Int64Value{Int64Value: v}}
}

func timeDistValue(bucketCounts []int64) *sc.MetricValue {
	return &sc.MetricValue{
		Value: &sc.MetricValue_DistributionValue{
			DistributionValue: &sc.Distribution{
				Count:        8,
				BucketCounts: bucketCounts,
				BucketOption: &sc.Distribution_ExponentialBuckets_{
					ExponentialBuckets: &sc.Distribution_ExponentialBuckets{
						NumFiniteBuckets: 8,
						GrowthFactor:     10,
						Scale:            1e-6,
					},
				},
			},
		},
	}
}

func sizeDistValue(bucketCounts []int64) *sc.MetricValue {
	return &sc.MetricValue{
		Value: &sc.MetricValue_DistributionValue{
			DistributionValue: &sc.Distribution{
				Count:        8,
				BucketCounts: bucketCounts,
				BucketOption: &sc.Distribution_ExponentialBuckets_{
					ExponentialBuckets: &sc.Distribution_ExponentialBuckets{
						NumFiniteBuckets: 8,
						GrowthFactor:     10,
						Scale:            1,
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

func TestConsume(t *testing.T) {
	var consumeTests = []struct {
		name    string
		statter *fakeStatter
		report  *sc.ReportRequest
		metrics map[string][]int64
	}{
		// TODO(dougreid): add more tests (ignored metrics, etc.)
		{
			name:    "Requests",
			statter: &fakeStatter{metrics: make(map[string][]int64)},
			report:  reportReq(svc, operation(shelfLbls, metricValueSet(mixologist.ProducerRequestCount, metricValue(14)))),
			metrics: map[string][]int64{
				"test-api-service-v1.appspot.com.us-east1.ListShelves.service.api.producer.request_count": []int64{14},
			},
		},
		{
			name:    "Requests -- Multiple MetricValues",
			statter: &fakeStatter{metrics: make(map[string][]int64)},
			report:  reportReq(svc, operation(shelfLbls, metricValueSet(mixologist.ProducerRequestCount, metricValue(14), metricValue(22)))),
			metrics: map[string][]int64{
				"test-api-service-v1.appspot.com.us-east1.ListShelves.service.api.producer.request_count": []int64{14, 22},
			},
		},
		{
			name:    "Total Latencies",
			statter: &fakeStatter{metrics: make(map[string][]int64)},
			report:  reportReq(svc, operation(shelfLbls, metricValueSet(mixologist.ProducerTotalLatencies, timeDistValue([]int64{1, 0, 1, 0, 0, 3, 0, 2, 0, 1})))),
			metrics: map[string][]int64{
				"test-api-service-v1.appspot.com.us-east1.ListShelves.service.api.producer.total_latencies": []int64{0, 0, 55, 55, 55, 5500, 5500, 50000},
			},
		},
		{
			name:    "Total Latencies -- Multiple",
			statter: &fakeStatter{metrics: make(map[string][]int64)},
			report:  reportReq(svc, operation(shelfLbls, metricValueSet(mixologist.ProducerTotalLatencies, timeDistValue([]int64{1, 0, 1, 0, 0, 3, 0, 2, 0, 1}), timeDistValue([]int64{0, 0, 0, 0, 5, 0, 1, 1, 3})))),
			metrics: map[string][]int64{
				"test-api-service-v1.appspot.com.us-east1.ListShelves.service.api.producer.total_latencies": []int64{0, 0, 55, 55, 55, 5500, 5500, 50000, 5, 5, 5, 5, 5, 550, 5500, 50000, 50000, 50000},
			},
		},
		{
			name:    "Request Sizes",
			statter: &fakeStatter{metrics: make(map[string][]int64)},
			report:  reportReq(svc, operation(bookLbls, metricValueSet(mixologist.ProducerRequestSizes, sizeDistValue([]int64{1, 0, 1, 0, 0, 3, 0, 2, 0, 1})))),
			metrics: map[string][]int64{
				"test-api-service-v1.appspot.com.us-east1.CreateBook.service.api.producer.request_sizes": []int64{1000, 55000, 55000000, 55000000, 55000000, 5500000000, 5500000000, 50000000000},
			},
		},
		{
			name:    "Mixed Metrics",
			statter: &fakeStatter{metrics: make(map[string][]int64)},
			report:  reportReq(svc, operation(bookLbls, metricValueSet(mixologist.ProducerRequestCount, metricValue(347)), metricValueSet(mixologist.ProducerRequestSizes, sizeDistValue([]int64{1, 0, 1, 0, 0, 3, 0, 2, 0, 1})))),
			metrics: map[string][]int64{
				"test-api-service-v1.appspot.com.us-east1.CreateBook.service.api.producer.request_sizes": []int64{1000, 55000, 55000000, 55000000, 55000000, 5500000000, 5500000000, 50000000000},
				"test-api-service-v1.appspot.com.us-east1.CreateBook.service.api.producer.request_count": []int64{347},
			},
		},
	}

	for _, v := range consumeTests {
		consumer := &consumer{client: v.statter}
		consumer.Consume([]*sc.ReportRequest{v.report})
		if eq := reflect.DeepEqual(v.metrics, v.statter.metrics); !eq {
			t.Errorf("%s: metrics not equal; got %v, want %v", v.name, v.statter.metrics, v.metrics)
		}
	}
}
