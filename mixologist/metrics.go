package mixologist

const (
	// Metrics names, etc. -- TODO(dougreid): refactor into metrics package
	ProducerRequestCount           = "serviceruntime.googleapis.com/api/producer/request_count"
	ProducerRequestSizes           = "serviceruntime.googleapis.com/api/producer/request_sizes"
	ProducerTotalLatencies         = "serviceruntime.googleapis.com/api/producer/total_latencies"
	ProducerBackendLatencies       = "serviceruntime.googleapis.com/api/producer/backend_latencies"
	ProducerRequestCountByConsumer = "serviceruntime.googleapis.com/api/producer/by_consumer/request_count"
	CloudProject                   = "cloud.googleapis.com/project"
	CloudLocation                  = "cloud.googleapis.com/location"
	CloudService                   = "cloud.googleapis.com/service"
	CloudUid                       = "cloud.googleapis.com/uid"
	APIVersion                     = "serviceruntime.googleapis.com/api_version"
	APIMethod                      = "serviceruntime.googleapis.com/api_method"
	ConsumerProject                = "serviceruntime.googleapis.com/consumer_project"
	Protocol                       = "/protocol"
	ResponseCode                   = "/response_code"
	ResponseCodeClass              = "/response_code_class"
	StatusCode                     = "/status_code"
	ConsumerID                     = "/consumer_id"
	CredentialID                   = "/credential_id"

	// Bucket details
	SizeDistributionScale        = 1
	SizeDistributionGrowthFactor = 10
	SizeDistributionBuckets      = 8

	TimeDistributionScale        = 1e-6
	TimeDistributionGrowthFactor = 10
	TimeDistributionBuckets      = 8
)

var (
	// Distributions
	SizeDistribution = &ExponentialDistribution{NumBuckets: SizeDistributionBuckets, StartValue: SizeDistributionScale, GrowthFactor: SizeDistributionGrowthFactor}
	TimeDistribution = &ExponentialDistribution{NumBuckets: TimeDistributionBuckets, StartValue: TimeDistributionScale, GrowthFactor: TimeDistributionGrowthFactor}

	// Labels
	MonitoredAPIResourceLabels = []string{CloudLocation, CloudUid, APIVersion, APIMethod, ConsumerProject, CloudProject, CloudService}
	PerMetricLabels            = map[string][]string{
		ProducerRequestCount:           []string{Protocol, ResponseCode, ResponseCodeClass, StatusCode},
		ProducerRequestCountByConsumer: []string{Protocol, ResponseCode, ResponseCodeClass, CredentialID, StatusCode},
	}
)

type ExponentialDistribution struct {
	NumBuckets   int32
	StartValue   float64
	GrowthFactor float64
}
