package fakes

import (
	"container/list"
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
)
