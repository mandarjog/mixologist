package mixologist

import (
	"github.com/golang/glog"
	"golang.org/x/net/context"
	sc "google/api/servicecontrol/v1"
)

// NewCheckerManager -- given a registry and a config object return a CheckerManager
func NewCheckerManager(registry map[string]CheckerBuilder, c Config) *CheckerManager {
	glog.Infof("creating checker manager with config: %v", c)
	checkerImpls := make([]Checker, 0, len(c.Checkers))
	for _, checkerName := range c.Checkers {
		if cn, ok := registry[checkerName]; ok {
			if cc, err := cn.BuildChecker(c); cc != nil {
				glog.Info("Built checker: ", checkerName, " ", cc)
				checkerImpls = append(checkerImpls, cc)
			} else {
				glog.Error("Unable to build checker: ", checkerName, " ", err)
			}
		}
	}
	glog.Info("Available Checkers: ", len(checkerImpls))
	return &CheckerManager{
		checkers: checkerImpls,
	}
}

// Check -- Top level check mehod that runs thru all registered checkers
func (c *CheckerManager) Check(ctx context.Context, msg *sc.CheckRequest) (*sc.CheckResponse, error) {
	// CheckErrors: []*sc.CheckError{&sc.CheckError{sc.CheckError_PERMISSION_DENIED, sc.CheckError_BUDGET_EXCEEDED.String()}},
	//ce := make([]*sc.CheckError, 0, len(c.checkers))
	ce := []*sc.CheckError{}
	for _, checker := range c.checkers {
		glog.V(1).Infof("Checking %s %s", checker.Name(), msg)
		cer, er := checker.Check(msg)
		if er != nil {
			cer = &sc.CheckError{sc.CheckError_PERMISSION_DENIED, er.Error()}
		}
		if cer != nil {
			ce = append(ce, cer)
		}
	}
	return &sc.CheckResponse{
		OperationId: msg.Operation.OperationId,
		CheckErrors: ce,
	}, nil
}
