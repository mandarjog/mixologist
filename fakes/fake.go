package fakes

import (
	"container/list"
	"fmt"
	sc "google/api/servicecontrol/v1"
	"net/http"
	"somnacin-internal/mixologist/mixologist"
	"sync"
)

// create a new builder
func NewBuilder(name string, meta map[string]interface{}) *builder {
	if meta == nil {
		meta = make(map[string]interface{})
	}
	bldr := &builder{
		name: name,
		meta: meta,
	}

	return bldr
}

// ReportConsumerBuilder
func (s *builder) NewConsumer(meta map[string]interface{}) mixologist.ReportConsumer {
	var prefixAndHandler *mixologist.PrefixAndHandler
	if meta == nil {
		meta = s.meta
	}
	if _prx, found := meta[HandlerPrefix]; found {
		prx := _prx.(string)
		prefixAndHandler = &mixologist.PrefixAndHandler{
			Prefix: prx,
			Handler: &handler{
				prefix: prx,
			},
		}

	}
	s.Consumer = &consumer{
		name:    s.name,
		meta:    meta,
		Msgs:    list.New(),
		handler: prefixAndHandler,
		lock:    &sync.Mutex{},
	}
	return s.Consumer
}

// ReportConsumer
func (s *consumer) GetName() string {
	return s.name
}

// a multi threaded consumer needs to take care of concurrency
func (s *consumer) Consume(reportMsg *sc.ReportRequest) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Msgs.PushBack(reportMsg)
	return nil
}

// returns messages that have been queued
func (s *consumer) GetMessages() []*sc.ReportRequest {
	s.lock.Lock()
	defer s.lock.Unlock()
	msgs := make([]*sc.ReportRequest, s.Msgs.Len())
	idx := 0
	for e := s.Msgs.Front(); e != nil; e = e.Next() {
		msgs[idx] = e.Value.(*sc.ReportRequest)
		idx++
	}
	return msgs
}

func (s *consumer) GetPrefixAndHandler() *mixologist.PrefixAndHandler {
	return s.handler
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello %s", h.prefix)
}
