package mixologist

import (
	"sync"
	"testing"
	"time"

	sc "google/api/servicecontrol/v1"
)

type fakeAdapter struct {
	ReportConsumer

	reqs       []*sc.ReportRequest
	numBatches int
	done       sync.WaitGroup
}

func (f *fakeAdapter) Consume(reqs []*sc.ReportRequest) error {
	f.reqs = append(f.reqs, reqs...)
	f.numBatches++
	f.done.Done()
	return nil
}

func TestConsume(t *testing.T) {
	tests := []struct {
		name              string
		in                []*sc.ReportRequest
		c                 BatchingConfig
		wantBatches       int
		closeAfterConsume bool
	}{
		{
			name: "Empty data",
			c:    BatchingConfig{},
		},
		{
			name:        "Single Request",
			in:          []*sc.ReportRequest{&sc.ReportRequest{ServiceName: "Test"}},
			c:           BatchingConfig{MaxBatchCount: 1, BatchTimeout: time.Minute},
			wantBatches: 1,
		},
		{
			name:        "Multiple Requests",
			in:          []*sc.ReportRequest{&sc.ReportRequest{ServiceName: "Test"}, &sc.ReportRequest{ServiceName: "Test 2"}, &sc.ReportRequest{ServiceName: "Test 3"}},
			c:           BatchingConfig{MaxBatchCount: 1, BatchTimeout: time.Minute},
			wantBatches: 3,
		},
		{
			name:        "Multiple Requests - Timeout",
			in:          []*sc.ReportRequest{&sc.ReportRequest{ServiceName: "Test"}, &sc.ReportRequest{ServiceName: "Test 2"}, &sc.ReportRequest{ServiceName: "Test 3"}},
			c:           BatchingConfig{BatchTimeout: time.Millisecond * 500},
			wantBatches: 1,
		},
	}

	for _, v := range tests {
		f := &fakeAdapter{numBatches: 0}

		f.done.Add(v.wantBatches)
		b := BatchingConsumer(f, v.c)
		b.Consume(v.in)
		f.done.Wait()

		if len(f.reqs) != len(v.in) {
			t.Errorf("%s: bad num of reqs; got %d, want %d", v.name, len(f.reqs), len(v.in))
		}
		if f.numBatches != v.wantBatches {
			t.Errorf("%s: bad num of batches; got %d, want %d", v.name, f.numBatches, v.wantBatches)
		}

		b.(*batcher).Close()
	}
}
