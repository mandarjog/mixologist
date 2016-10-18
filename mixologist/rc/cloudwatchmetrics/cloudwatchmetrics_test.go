package cloudwatchmetrics

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/cloudendpoints/mixologist/mixologist"

	timestamppb "github.com/golang/protobuf/ptypes/timestamp"
	servicecontrol "google/api/servicecontrol/v1"
)

type fakeAPI struct {
	cloudwatchiface.CloudWatchAPI

	returnErr bool
	lastReq   *cloudwatch.PutMetricDataInput
}

type Dimensions []*cloudwatch.Dimension

func (d Dimensions) Len() int           { return len(d) }
func (d Dimensions) Less(i, j int) bool { return *d[i].Name < *d[j].Name }
func (d Dimensions) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }

func (f *fakeAPI) PutMetricData(in *cloudwatch.PutMetricDataInput) (*cloudwatch.PutMetricDataOutput, error) {
	f.lastReq = in
	return nil, nil
}

var (
	svc       = "test-svc"
	testTime  = time.Unix(9009, 0)
	shelfLbls = map[string]string{
		mixologist.CloudService:  svc,
		mixologist.CloudLocation: "us-east1",
		mixologist.APIVersion:    "test-api-service-v1.appspot.com",
		mixologist.APIMethod:     "ListShelves",
	}
)

func metricValue(v int64) *servicecontrol.MetricValue {
	return &servicecontrol.MetricValue{Value: &servicecontrol.MetricValue_Int64Value{Int64Value: v}}
}

func distValue(count int64, min, max, mean float64) *servicecontrol.MetricValue {
	return &servicecontrol.MetricValue{
		Value: &servicecontrol.MetricValue_DistributionValue{
			DistributionValue: &servicecontrol.Distribution{
				Count:   count,
				Mean:    mean,
				Minimum: min,
				Maximum: max,
			},
		},
	}
}

func metricValueSet(name string, mv ...*servicecontrol.MetricValue) *servicecontrol.MetricValueSet {
	return &servicecontrol.MetricValueSet{MetricName: name, MetricValues: mv}
}

func operation(labels map[string]string, mvs ...*servicecontrol.MetricValueSet) *servicecontrol.Operation {
	return &servicecontrol.Operation{
		Labels:          labels,
		StartTime:       &timestamppb.Timestamp{Seconds: 9009},
		MetricValueSets: mvs,
	}
}

func reportReq(name string, ops ...*servicecontrol.Operation) *servicecontrol.ReportRequest {
	return &servicecontrol.ReportRequest{ServiceName: name, Operations: ops}
}

func TestConsume(t *testing.T) {
	var consumeTests = []struct {
		name   string
		report *servicecontrol.ReportRequest
		want   *cloudwatch.PutMetricDataInput
	}{
		{
			name:   "Requests",
			report: reportReq(svc, operation(shelfLbls, metricValueSet(mixologist.ProducerRequestCount, metricValue(14)))),
			want: &cloudwatch.PutMetricDataInput{
				MetricData: []*cloudwatch.MetricDatum{
					&cloudwatch.MetricDatum{
						MetricName: aws.String(mixologist.ProducerRequestCount),
						Dimensions: []*cloudwatch.Dimension{
							&cloudwatch.Dimension{Name: aws.String(mixologist.CloudLocation), Value: aws.String("us-east1")},
							&cloudwatch.Dimension{Name: aws.String(mixologist.CloudService), Value: aws.String(svc)},
							&cloudwatch.Dimension{Name: aws.String(mixologist.APIMethod), Value: aws.String("ListShelves")},
							&cloudwatch.Dimension{Name: aws.String(mixologist.APIVersion), Value: aws.String("test-api-service-v1.appspot.com")},
						},
						Timestamp: aws.Time(testTime),
						Value:     aws.Float64(14),
					},
				},
				Namespace: aws.String("mixologist.io"),
			},
		},
		{
			name:   "Requests -- Multiple MetricValues",
			report: reportReq(svc, operation(shelfLbls, metricValueSet(mixologist.ProducerRequestCount, metricValue(14), metricValue(22)))),
			want: &cloudwatch.PutMetricDataInput{
				MetricData: []*cloudwatch.MetricDatum{
					&cloudwatch.MetricDatum{
						MetricName: aws.String(mixologist.ProducerRequestCount),
						Dimensions: []*cloudwatch.Dimension{
							&cloudwatch.Dimension{Name: aws.String(mixologist.CloudLocation), Value: aws.String("us-east1")},
							&cloudwatch.Dimension{Name: aws.String(mixologist.CloudService), Value: aws.String(svc)},
							&cloudwatch.Dimension{Name: aws.String(mixologist.APIMethod), Value: aws.String("ListShelves")},
							&cloudwatch.Dimension{Name: aws.String(mixologist.APIVersion), Value: aws.String("test-api-service-v1.appspot.com")},
						},
						Timestamp: aws.Time(testTime),
						Value:     aws.Float64(14),
					},
					&cloudwatch.MetricDatum{
						MetricName: aws.String(mixologist.ProducerRequestCount),
						Dimensions: []*cloudwatch.Dimension{
							&cloudwatch.Dimension{Name: aws.String(mixologist.CloudLocation), Value: aws.String("us-east1")},
							&cloudwatch.Dimension{Name: aws.String(mixologist.CloudService), Value: aws.String(svc)},
							&cloudwatch.Dimension{Name: aws.String(mixologist.APIMethod), Value: aws.String("ListShelves")},
							&cloudwatch.Dimension{Name: aws.String(mixologist.APIVersion), Value: aws.String("test-api-service-v1.appspot.com")},
						},
						Timestamp: aws.Time(testTime),
						Value:     aws.Float64(22),
					},
				},
				Namespace: aws.String("mixologist.io"),
			},
		},
		{
			name:   "Total Latencies",
			report: reportReq(svc, operation(shelfLbls, metricValueSet(mixologist.ProducerTotalLatencies, distValue(8, 0.0756, .756, 0.345)))),
			want: &cloudwatch.PutMetricDataInput{
				MetricData: []*cloudwatch.MetricDatum{
					&cloudwatch.MetricDatum{
						MetricName: aws.String(mixologist.ProducerTotalLatencies),
						Dimensions: []*cloudwatch.Dimension{
							&cloudwatch.Dimension{Name: aws.String(mixologist.CloudLocation), Value: aws.String("us-east1")},
							&cloudwatch.Dimension{Name: aws.String(mixologist.CloudService), Value: aws.String(svc)},
							&cloudwatch.Dimension{Name: aws.String(mixologist.APIMethod), Value: aws.String("ListShelves")},
							&cloudwatch.Dimension{Name: aws.String(mixologist.APIVersion), Value: aws.String("test-api-service-v1.appspot.com")},
						},
						Timestamp: aws.Time(testTime),
						StatisticValues: &cloudwatch.StatisticSet{
							Maximum:     aws.Float64(0.756),
							Minimum:     aws.Float64(0.0756),
							SampleCount: aws.Float64(8),
							Sum:         aws.Float64(2.76),
						},
					},
				},
				Namespace: aws.String("mixologist.io"),
			},
		},
		{
			name:   "Mixed Metrics",
			report: reportReq(svc, operation(shelfLbls, metricValueSet(mixologist.ProducerRequestCount, metricValue(347)), metricValueSet(mixologist.ProducerTotalLatencies, distValue(16, 0.0034, 5.764, .893)))),
			want: &cloudwatch.PutMetricDataInput{
				MetricData: []*cloudwatch.MetricDatum{
					&cloudwatch.MetricDatum{
						MetricName: aws.String(mixologist.ProducerRequestCount),
						Dimensions: []*cloudwatch.Dimension{
							&cloudwatch.Dimension{Name: aws.String(mixologist.CloudLocation), Value: aws.String("us-east1")},
							&cloudwatch.Dimension{Name: aws.String(mixologist.CloudService), Value: aws.String(svc)},
							&cloudwatch.Dimension{Name: aws.String(mixologist.APIMethod), Value: aws.String("ListShelves")},
							&cloudwatch.Dimension{Name: aws.String(mixologist.APIVersion), Value: aws.String("test-api-service-v1.appspot.com")},
						},
						Timestamp: aws.Time(testTime),
						Value:     aws.Float64(347),
					},

					&cloudwatch.MetricDatum{
						MetricName: aws.String(mixologist.ProducerTotalLatencies),
						Dimensions: []*cloudwatch.Dimension{
							&cloudwatch.Dimension{Name: aws.String(mixologist.CloudLocation), Value: aws.String("us-east1")},
							&cloudwatch.Dimension{Name: aws.String(mixologist.CloudService), Value: aws.String(svc)},
							&cloudwatch.Dimension{Name: aws.String(mixologist.APIMethod), Value: aws.String("ListShelves")},
							&cloudwatch.Dimension{Name: aws.String(mixologist.APIVersion), Value: aws.String("test-api-service-v1.appspot.com")},
						},
						Timestamp: aws.Time(testTime),
						StatisticValues: &cloudwatch.StatisticSet{
							Maximum:     aws.Float64(5.764),
							Minimum:     aws.Float64(0.0034),
							SampleCount: aws.Float64(16),
							Sum:         aws.Float64(16 * .893),
						},
					},
				},
				Namespace: aws.String("mixologist.io"),
			},
		},
	}

	for _, v := range consumeTests {
		f := &fakeAPI{} //returnErr: v.reportErr}
		consumer := &consumer{cw: f}
		consumer.Consume([]*servicecontrol.ReportRequest{v.report})

		// predictable ordering of dimension data allows for better reflect.DeepEqual comparison.
		for _, v := range f.lastReq.MetricData {
			sort.Sort(Dimensions(v.Dimensions))
		}

		if !reflect.DeepEqual(f.lastReq, v.want) {
			t.Errorf("%s: bad request generated; got %v, want %v", v.name, f.lastReq, v.want)
		}
	}
}
