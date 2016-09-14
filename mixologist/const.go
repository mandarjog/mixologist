package mixologist

const (
	// Port -- Default server port
	Port = 9092
	// NConsumers -- number of consumer threads
	NConsumers = 2
	// CheckSuffix -- to identify a POST request as check
	CheckSuffix = ":check"
	// ReportSuffix -- to identify a POST request as report
	ReportSuffix = ":report"
)
