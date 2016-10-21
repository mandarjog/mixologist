// Package cloudwatchmetrics enables metrics export to AWS CloudWatch.
//
// To use this package, one must currently create a credentials.ini file that contains
// the following:
//
// [default]
// aws_access_key_id=<KEY>
// aws_secret_access_key=<SECRET>
//
// Then, one must create a kubernetes secret in the namespace that mixologist will be
// run in. Then the mixologist RC must be modified to mount the volume to /etc/aws.
//
// In the future, there will likely be a better, more dynamic way of passing in creds for
// use in this package.`
package cloudwatchmetrics

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/cloudendpoints/mixologist/mixologist"
	"github.com/golang/glog"

	servicecontrol "google/api/servicecontrol/v1"
)

const (
	// Name is the unique identifier for this adapter
	Name = "aws/cloudwatchmetrics"

	// Region selects the AWS region (and endpoint) to send logs.
	// TODO: make a configuration option
	Region = "us-west-2" // US West (Oregon) us-west-2 logs.us-west-2.amazonaws.com HTTPS

	// creds details (const for now -- could move into supplier)
	svcAcctFile    = "/etc/aws/credentials.ini"
	svcAcctProfile = "default"

	// used for adding unit values to distribution metrics
	// Valid Values: Seconds | Microseconds | Milliseconds | Bytes | Kilobytes | Megabytes | Gigabytes | Terabytes | Bits | Kilobits | Megabits | Gigabits | Terabits | Percent | Count | Bytes/Second | Kilobytes/Second | Megabytes/Second | Gigabytes/Second | Terabytes/Second | Bits/Second | Kilobits/Second | Megabits/Second | Gigabits/Second | Terabits/Second | Count/Second | None
	timeUnit = "Seconds"
	sizeUnit = "Bytes"
)

var (
	// AllowedDimensions controls the labels from ESP that are allowed for export to
	// cloudwatch. These prevents the inclusion of too many dimensions for metrics.
	AllowedDimensions = map[string]struct{}{
		mixologist.APIMethod:         struct{}{},
		mixologist.APIVersion:        struct{}{},
		mixologist.CloudService:      struct{}{},
		mixologist.CloudLocation:     struct{}{},
		mixologist.ConsumerID:        struct{}{},
		mixologist.CredentialID:      struct{}{},
		mixologist.Protocol:          struct{}{},
		mixologist.ResponseCode:      struct{}{},
		mixologist.ResponseCodeClass: struct{}{},
		mixologist.StatusCode:        struct{}{},
	}
)

func init() {
	mixologist.RegisterReportConsumer(Name, new(builder))
}

type (
	consumer struct {
		cw cloudwatchiface.CloudWatchAPI
	}

	builder struct{}
)

func timestamp(m *servicecontrol.MetricValue, o *servicecontrol.Operation) *time.Time {

	// order of preference:
	// (1) metric start time
	// (2) metric end time
	// (3) operation start time
	// (4) operation end time
	// (5) current time

	ts := m.StartTime
	if ts == nil {
		ts = m.EndTime
	}
	if ts == nil {
		ts = o.GetStartTime()
	}
	if ts == nil {
		ts = o.GetEndTime()
	}
	if ts == nil {
		return aws.Time(time.Now())
	}

	return aws.Time(time.Unix(ts.Seconds, int64(ts.Nanos)))
}

// TODO: take care to only allow 10 dimensions
// Metrics may have up to 10 dimensions
func dimensions(labels map[string]string) []*cloudwatch.Dimension {
	var d []*cloudwatch.Dimension
	for k, v := range labels {
		if v != "" {
			d = append(d, &cloudwatch.Dimension{Name: aws.String(k), Value: aws.String(v)})
		}
	}
	return d
}

func filter(d []*cloudwatch.Dimension) []*cloudwatch.Dimension {
	var filtered []*cloudwatch.Dimension
	for _, di := range d {
		if _, found := AllowedDimensions[aws.StringValue(di.Name)]; found {
			filtered = append(filtered, di)
		}
	}

	// TODO: ensure total size <= 10

	return filtered
}

func singleValueMetric(name string, m *servicecontrol.MetricValue, o *servicecontrol.Operation) *cloudwatch.MetricDatum {
	return &cloudwatch.MetricDatum{
		MetricName: aws.String(name),
		Value:      aws.Float64(float64(m.GetInt64Value())),
		Dimensions: filter(append(dimensions(o.GetLabels()), dimensions(m.GetLabels())...)),
		Timestamp:  timestamp(m, o),
	}
}

func stats(m *servicecontrol.MetricValue) *cloudwatch.StatisticSet {
	d := m.GetDistributionValue()
	if d == nil {
		return nil
	}

	// TODO: for now, assume these are fully populated.
	count := float64(d.Count)
	max := d.Maximum
	min := d.Minimum
	sum := d.Mean * count

	sv := &cloudwatch.StatisticSet{
		SampleCount: aws.Float64(count),
		Maximum:     aws.Float64(max),
		Minimum:     aws.Float64(min),
		Sum:         aws.Float64(sum),
	}

	return sv
}

func distributionMetric(name string, m *servicecontrol.MetricValue, o *servicecontrol.Operation) *cloudwatch.MetricDatum {
	return &cloudwatch.MetricDatum{
		MetricName:      aws.String(name),
		Dimensions:      filter(append(dimensions(o.GetLabels()), dimensions(m.GetLabels())...)),
		Timestamp:       timestamp(m, o),
		StatisticValues: stats(m),
	}
}

func (c *consumer) Publish(m []*cloudwatch.MetricDatum) error {

	// TODO:
	// Each PutMetricData request is limited to 8 KB in size for HTTP GET requests
	// and 40 KB for HTTP PUT requests

	params := &cloudwatch.PutMetricDataInput{
		MetricData: m,
		Namespace:  aws.String("mixologist.io"), // Required
	}

	glog.Infof("publishing metrics to aws: %v", params)

	_, err := c.cw.PutMetricData(params)
	if err != nil {
		glog.Errorf("error publishing metric data: %v", err)
		return err
	}
	return nil
}

func (c *consumer) process(o *servicecontrol.Operation) error {
	var metrics []*cloudwatch.MetricDatum
	for _, mvs := range o.GetMetricValueSets() {
		for _, mv := range mvs.GetMetricValues() {
			switch mv.Value.(type) {
			case *servicecontrol.MetricValue_Int64Value:
				metrics = append(metrics, singleValueMetric(mvs.MetricName, mv, o))
			case *servicecontrol.MetricValue_DistributionValue:
				metrics = append(metrics, distributionMetric(mvs.MetricName, mv, o))
			}
		}
	}
	return c.Publish(metrics)
}

// Consume -- Called to consume multiple reportMsgs (batch support)
func (c *consumer) Consume(reportMsgs []*servicecontrol.ReportRequest) error {
	for _, reportMsg := range reportMsgs {
		// svc := reportMsg.ServiceName
		for _, oprn := range reportMsg.GetOperations() {

			// TODO: better error handling
			c.process(oprn)
		}
	}

	// TODO: return error when it makes sense
	return nil
}

// GetName interface method
func (c *consumer) GetName() string {
	return Name
}

// GetPrefixAndHandler is not implemented by this adapter.
func (c *consumer) GetPrefixAndHandler() *mixologist.PrefixAndHandler {
	return nil
}

// BuildConsumer builds the adapter based on the configuration options passed in via the
// config struct.
func (b *builder) BuildConsumer(c mixologist.Config) (cc mixologist.ReportConsumer, err error) {

	creds := credentials.NewSharedCredentials(svcAcctFile, svcAcctProfile)
	config := aws.NewConfig().WithCredentials(creds).WithRegion(Region)
	sess, err := session.NewSession(config)
	if err != nil {
		glog.Errorf("could not create aws cloudwatchlogs session: %v", err)
		// TODO: build better error handling here
		return nil, err
	}
	svc := cloudwatch.New(sess)
	return &consumer{cw: svc}, nil
}
