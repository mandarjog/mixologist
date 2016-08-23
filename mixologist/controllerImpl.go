package mixologist

import (
	"encoding/json"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	sc "google/api/servicecontrol/v1"
)

// Check implementation
// Always return a success
func (c *ControllerImpl) Check(ctx context.Context, msg *sc.CheckRequest) (*sc.CheckResponse, error) {
	resp := &sc.CheckResponse{
		OperationId: msg.Operation.OperationId,
		// CheckErrors: []*sc.CheckError{&sc.CheckError{sc.CheckError_PERMISSION_DENIED, sc.CheckError_BUDGET_EXCEEDED.String()}},
		CheckErrors: []*sc.CheckError{},
	}
	return resp, nil
}

// Report into a log file
func (c *ControllerImpl) Report(ctx context.Context, msg *sc.ReportRequest) (*sc.ReportResponse, error) {
	dbg, _ := json.Marshal(msg)
	c.ReportQueue <- *msg
	glog.Info(string(dbg))
	resp := &sc.ReportResponse{}
	return resp, nil
}

func NewControllerImpl() *ControllerImpl {
	return &ControllerImpl{
		ReportQueue: make(chan sc.ReportRequest),
	}
}
