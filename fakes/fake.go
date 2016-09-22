package fakes

import (
	"container/list"
	"crypto/rand"
	"fmt"
	"golang.org/x/net/context"
	sc "google/api/servicecontrol/v1"
	"net/http"
	"somnacin-internal/mixologist/mixologist"
	"sync"
)

// create a new builder
func NewBuilder(name string, err error) *builder {
	bldr := &builder{
		name: name,
		err:  err,
	}
	return bldr
}

func BuildPrefixAndHandler(prx string) *mixologist.PrefixAndHandler {
	return &mixologist.PrefixAndHandler{
		Prefix: prx,
		Handler: &handler{
			prefix: prx,
		},
	}

}

// ReportConsumerBuilder
func (s *builder) BuildConsumer(c mixologist.Config) (mixologist.ReportConsumer, error) {
	s.Consumer = &consumer{
		name:    s.name,
		Msgs:    list.New(),
		handler: BuildPrefixAndHandler("fake-handler"),
		lock:    &sync.Mutex{},
	}
	return s.Consumer, s.err
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

func UUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
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
	fmt.Fprintf(w, "%s", h.prefix)
}

// Check implementation
// Always return a success
func (c *controller) Check(ctx context.Context, msg *sc.CheckRequest) (*sc.CheckResponse, error) {
	resp := &sc.CheckResponse{
		OperationId: msg.Operation.OperationId,
	}
	c.SpyCR = msg
	return resp, c.PlantedError
}

// Report into a log file
func (c *controller) Report(ctx context.Context, msg *sc.ReportRequest) (*sc.ReportResponse, error) {
	c.SpyRR = msg
	return &sc.ReportResponse{}, c.PlantedError
}

func (c *controller) GetReportQueue() chan *sc.ReportRequest {
	return c.reportQueue
}

// build a controller
func NewController() *controller {
	return &controller{reportQueue: make(chan *sc.ReportRequest)}
}
