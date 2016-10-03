package mixologist

import (
	"golang.org/x/net/context"
	sc "google/api/servicecontrol/v1"
)

// Check implementation
func (c *ControllerImpl) Check(ctx context.Context, msg *sc.CheckRequest) (*sc.CheckResponse, error) {
	return c.checkerManager.Check(ctx, msg)
}

// Report into a log file
func (c *ControllerImpl) Report(ctx context.Context, msg *sc.ReportRequest) (*sc.ReportResponse, error) {
	c.reportQueue <- msg
	resp := &sc.ReportResponse{}
	return resp, nil
}

// ReportQueue -- get a reference to the underlying channel
func (c *ControllerImpl) ReportQueue() chan *sc.ReportRequest {
	return c.reportQueue
}

// NewControllerImpl - return a newly created controller
func NewControllerImpl(cm *CheckerManager) Controller {
	return &ControllerImpl{
		reportQueue:    make(chan *sc.ReportRequest),
		checkerManager: cm,
	}
}
