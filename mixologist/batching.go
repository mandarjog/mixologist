package mixologist

import (
	"fmt"
	"time"

	sc "google/api/servicecontrol/v1"
)

const (
	DefaultBatchMaxCount = 50
	DefaultBatchTimeout  = time.Minute
)

var (
	defaultConfig = BatchingConfig{
		MaxBatchCount: DefaultBatchMaxCount,
		BatchTimeout:  DefaultBatchTimeout,
	}
)

type batcher struct {
	bufChan  chan *sc.ReportRequest
	closing  chan struct{}
	consumer ReportConsumer
}

type BatchingConfig struct {
	MaxBatchCount int
	BatchTimeout  time.Duration
}

func (b *batcher) flush(reqs []*sc.ReportRequest) int {
	if len(reqs) > 0 {
		b.consumer.Consume(reqs)
	}
	return len(reqs)
}

func (b *batcher) GetName() string {
	return fmt.Sprintf("Batching Adapter for: %s", b.consumer.GetName())
}

func (b *batcher) GetPrefixAndHandler() *PrefixAndHandler {
	return b.consumer.GetPrefixAndHandler()
}

func (b *batcher) Consume(reqs []*sc.ReportRequest) error {
	for _, req := range reqs {
		b.bufChan <- req
	}
	return nil
}

// TOOD: not yet a method on ReportConsumer, but we probably want to add it
func (b *batcher) Close() {
	close(b.closing)
	// TODO: should close bufChan? signal Consume to start reporting errors?
	// close(b.bufChan)
}

func (b *batcher) batchLoop(max int, timeout time.Duration) {
	t := time.NewTicker(timeout)
	defer t.Stop()

	for {
		var reqs []*sc.ReportRequest
		batchFull := false
		for !batchFull {
			select {
			case req := <-b.bufChan:
				reqs = append(reqs, req)
				if len(reqs) >= max {
					batchFull = true
				}
			case <-t.C:
				batchFull = true
			case <-b.closing:
				b.flush(reqs)
				return
			}
		}
		b.flush(reqs)
	}
}

func BatchingConsumer(consumer ReportConsumer, c BatchingConfig) ReportConsumer {

	conf := defaultConfig
	if c.MaxBatchCount > 0 {
		conf.MaxBatchCount = c.MaxBatchCount
	}
	if c.BatchTimeout > 0 {
		conf.BatchTimeout = c.BatchTimeout
	}

	b := &batcher{
		consumer: consumer,
		bufChan:  make(chan *sc.ReportRequest, conf.MaxBatchCount),
		closing:  make(chan struct{}, 1),
	}

	go b.batchLoop(conf.MaxBatchCount, conf.BatchTimeout)

	return b
}
