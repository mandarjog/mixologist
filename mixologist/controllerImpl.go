package mixologist

import (
	"golang.org/x/net/context"
	sc "google/api/servicecontrol/v1"
)

// Check implementation
// Always return a success
func (c *ControllerImpl) Check(ctx context.Context, msg *sc.CheckRequest) (*sc.CheckResponse, error) {
	resp := &sc.CheckResponse{
		OperationId: msg.Operation.OperationId,
		// CheckErrors: []*sc.CheckError{&sc.CheckError{sc.CheckError_PERMISSION_DENIED, sc.CheckError_BUDGET_EXCEEDED.String()}},
	}
	return resp, nil
}

// Report into a log file
func (c *ControllerImpl) Report(ctx context.Context, msg *sc.ReportRequest) (*sc.ReportResponse, error) {
	c.ReportQueue <- msg
	resp := &sc.ReportResponse{}
	return resp, nil
}

// NewControllerImpl - return a newly created controller
func NewControllerImpl() *ControllerImpl {
	return &ControllerImpl{
		ReportQueue: make(chan *sc.ReportRequest),
	}
}
