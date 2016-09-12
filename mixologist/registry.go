package mixologist

import (
	"errors"
)

var (
	// ReportConsumerRegistry -- all reportconsumerbuilders are registered here
	ReportConsumerRegistry = make(map[string]ReportConsumerBuilder)
)

// RegisterReportConsumer -- all report consumer should register their builder here
// This is typically done in the init() of the package rc/${consumer}/
func RegisterReportConsumer(name string, builder ReportConsumerBuilder) error {
	if val, collision := ReportConsumerRegistry[name]; collision {
		if val != builder {
			return errors.New("Collision")
		}
	} else {
		ReportConsumerRegistry[name] = builder
	}
	return nil
}
