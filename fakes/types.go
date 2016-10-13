package fakes

import (
	"container/list"
	"github.com/cloudendpoints/mixologist/mixologist"
	sc "google/api/servicecontrol/v1"
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
		err      error
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

	checker struct {
		name string
		meta map[string]interface{}
		Msgs *list.List
	}
	checkerbuilder struct {
		name    string
		err     error
		meta    map[string]interface{}
		Checker *checker
	}
	flist struct {
		Wl string `yaml:",omitempty,required"`
	}
	checkerConfig struct {
		OnCall string `yaml:",omitempty,required"`
		Flist  flist
	}
)
