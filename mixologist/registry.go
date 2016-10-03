package mixologist

import (
	"errors"
)

var (
	// ReportConsumerRegistry -- all reportconsumerbuilders are registered here
	ReportConsumerRegistry = make(map[string]ReportConsumerBuilder)
	// CheckerRegistry -- all checker are registered here
	CheckerRegistry = make(map[string]CheckerBuilder)
)

type (
	getter func() (interface{}, bool)
	setter func()
)

// RegisterReportConsumer -- all report consumers should register their builder here
// This is typically done in the init() of the package rc/${consumer}/
func RegisterReportConsumer(name string, builder ReportConsumerBuilder) error {
	return register(name, builder, func() (interface{}, bool) {
		val, collision := ReportConsumerRegistry[name]
		return val, collision
	}, func() {
		ReportConsumerRegistry[name] = builder
	})
}

// RegisterChecker -- all checkers should register their builder here
// This is typically done in the init() of the package rc/${consumer}/
func RegisterChecker(name string, builder CheckerBuilder) error {
	return register(name, builder, func() (interface{}, bool) {
		val, collision := CheckerRegistry[name]
		return val, collision
	}, func() {
		CheckerRegistry[name] = builder
	})
}

// abstracts collision or any other logic
func register(name string, builder interface{}, get getter, set setter) error {
	if val, collision := get(); collision {
		if val != builder {
			return errors.New("Collision")
		}
	} else {
		set()
	}
	return nil
}
