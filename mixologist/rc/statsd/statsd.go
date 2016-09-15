package statsd

import (
	sd "github.com/cactus/go-statsd-client/statsd"
	"github.com/golang/glog"
	sc "google/api/servicecontrol/v1"
	"math/big"
	"somnacin-internal/mixologist/mixologist"
	"strings"
	"time"
)

var (
	sizeHistogramBuckets = []float64{1, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7}
	timeHistogramBuckets = []float64{1e-6, 1e-5, 1e-4, 1e-3, 1e-2, 1e-1, 1, 1e1}
	metricNameReplacer   = strings.NewReplacer("serviceruntime.googleapis.com", "service", "cloud.googleapis.com", "cloud", "/", ".")
)

func resourcePrefix(labels map[string]string) string {
	pieces := []string{}
	if ver, ok := labels[mixologist.APIVersion]; ok {
		pieces = append(pieces, ver)
	}
	if loc, ok := labels[mixologist.CloudLocation]; ok {
		pieces = append(pieces, loc)
	}
	if meth, ok := labels[mixologist.APIMethod]; ok {
		pieces = append(pieces, meth)
	}
	return strings.Trim(strings.Join(pieces, "."), ".")
}

func metricSuffix(metric string, labels map[string]string) string {
	s := []string{}
	// TODO(dougreid): sort to eliminate worry about ordering
	for _, k := range mixologist.PerMetricLabels[metric] {
		if v, ok := labels[k]; ok {
			s = append(s, v)
		}
	}
	return strings.Trim(strings.Join(s, "."), ".")
}

func metricName(prefix, name, suffix string) string {
	return strings.Trim(strings.Join([]string{prefix, metricNameReplacer.Replace(name), suffix}, "."), ".")
}

// TODO(dougreid): extract to metrics.go
func bucketsMatch(b *sc.Distribution_ExponentialBuckets, d *mixologist.ExponentialDistribution) bool {
	if b == nil {
		return false
	}
	return b.NumFiniteBuckets == d.NumBuckets &&
		big.NewFloat(b.GrowthFactor).Cmp(big.NewFloat(d.GrowthFactor)) == 0 &&
		big.NewFloat(b.Scale).Cmp(big.NewFloat(d.StartValue)) == 0
}

func (c *consumer) populateFromBuckets(n string, dist *sc.Distribution, ed *mixologist.ExponentialDistribution, buckets []float64) {
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
			// convert to int64 millisecond value (all that is supported by statsd)
			// this will lead to a bunch of 0s (TODO(dougreid): should they be filtered out?)
			val := float64(curr) * (float64(time.Second) / float64(time.Millisecond))
			if err := c.client.Timing(n, int64(val), 1.0); err != nil {
				glog.Errorf("could not write timing value: %v", err)
			}
		}
	}
}

func (c *consumer) update(mv *sc.MetricValue, scName, metric string) {
	d := mv.GetDistributionValue()
	if d == nil {
		return
	}
	switch scName {
	case mixologist.ProducerTotalLatencies, mixologist.ProducerBackendLatencies:
		// time buckets
		c.populateFromBuckets(metric, d, mixologist.TimeDistribution, timeHistogramBuckets)
	case mixologist.ProducerRequestSizes:
		// size buckets
		c.populateFromBuckets(metric, d, mixologist.SizeDistribution, sizeHistogramBuckets)
	default:
		glog.Warningf("unknown metric for distribution: %s", metric)
	}
}

func (c *consumer) process(mvs *sc.MetricValueSet, prefix string) {
	for _, mv := range mvs.GetMetricValues() {

		switch mv.Value.(type) {
		case *sc.MetricValue_Int64Value:
			// counter
			n := metricName(prefix, mvs.MetricName, metricSuffix(mvs.MetricName, mv.Labels))

			if err := c.client.Inc(n, mv.GetInt64Value(), 1.0); err != nil {
				glog.Errorf("could not update statsd counter: %v", err)
			}
		case *sc.MetricValue_DistributionValue:
			n := metricName(prefix, mvs.MetricName, "")
			c.update(mv, mvs.MetricName, n)
		}
	}
}

// Consume -- Called to consume 1 reportMsg at a time
func (c *consumer) Consume(reportMsg *sc.ReportRequest) error {
	svc := reportMsg.ServiceName
	for _, oprn := range reportMsg.GetOperations() {
		defaultLabels := oprn.GetLabels()
		defaultLabels[mixologist.CloudService] = svc
		defaultLabels[mixologist.ConsumerID] = oprn.ConsumerId

		pre := resourcePrefix(defaultLabels)

		for _, mvs := range oprn.GetMetricValueSets() {
			c.process(mvs, pre)
		}
	}

	//TODO return error when it makes sense
	return nil
}

// GetName interface method
func (c *consumer) GetName() string {
	return Name
}

// Not needed.
func (c *consumer) GetPrefixAndHandler() *mixologist.PrefixAndHandler {
	return nil
}

func (b *builder) New(meta map[string]interface{}) mixologist.ReportConsumer {
	if c, err := sd.NewClient(meta[AddrArg].(string), ""); err == nil {
		return &consumer{
			client: c,
		}
	}
	return nil
}
