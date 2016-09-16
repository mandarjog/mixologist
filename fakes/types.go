package fakes

import (
	"container/list"
	sc "google/api/servicecontrol/v1"
	"somnacin-internal/mixologist/mixologist"
	"sync"
)

type (
	consumer struct {
		name    string
		meta    map[string]interface{}
		Msgs    *list.List
		handler *mixologist.PrefixAndHandler
		lock    *sync.Mutex
	}
	builder struct {
		name     string
		meta     map[string]interface{}
		Consumer *consumer
	}

	handler struct {
		prefix string
	}
	controller struct {
		reportQueue  chan *sc.ReportRequest
		SpyRR        *sc.ReportRequest
		SpyCR        *sc.CheckRequest
		PlantedError error
	}
)