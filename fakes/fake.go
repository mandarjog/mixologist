package fakes

import (
	sc "google/api/servicecontrol/v1"
	"somnacin-internal/mixologist/mixologist"
)

func NewBuilder(name string) mixologist.ReportConsumerBuilder {
	return &builder{
		name: name,
	}
}

// ReportConsumerBuilder
func (s *builder) New(meta map[string]interface{}) mixologist.ReportConsumer {
	return &consumer{
		name: s.name,
		meta: meta,
		msgs: make([]*sc.ReportRequest, 5),
	}
}

// ReportConsumer
func (s *consumer) GetName() string {
	return s.name
}

func (s *consumer) Consume(reportMsg *sc.ReportRequest) error {
	return nil
}

func (s *consumer) GetPrefixAndHandler() *mixologist.PrefixAndHandler {
	return nil
}
