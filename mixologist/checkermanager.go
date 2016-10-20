package mixologist

import (
	sc "google/api/servicecontrol/v1"

	"github.com/golang/glog"
	"golang.org/x/net/context"
)

// NewCheckerManager -- given a registry and a config object return a CheckerManager
func NewCheckerManager(registry map[string]CheckerBuilder, cfg *ServicesConfig) (*CheckerManager, []error) {
	var erra []error
	cm := &CheckerManager{
		checkers: make(map[ConstructorParams]Checker),
	}
	cm.cfg.Store(cfg)

	return cm, erra
}

// FindChecker -- given a checker kind and AdapterParams return a checker
func (c *CheckerManager) FindChecker(kind string, ru *RuntimeAdapterState) (chk Checker, err error) {
	key := ConstructorParams{
		Kind:   kind,
		Params: ru.TypedParams,
	}
	c.lock.RLock()
	chk, found := c.checkers[key]
	c.lock.RUnlock()
	if found {
		return chk, nil
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	chk, found = c.checkers[key]
	if !found {
		if chk, err = ru.Builder.(CheckerBuilder).BuildChecker(ru.TypedParams); err != nil {
			return nil, err
		}
		c.checkers[key] = chk
	}
	return chk, nil
}

// Check -- Top level check mehod that runs thru all registered checkers
func (c *CheckerManager) Check(ctx context.Context, msg *sc.CheckRequest) (*sc.CheckResponse, error) {
	cfg := c.cfg.Load().(*ServicesConfig)
	checkers := cfg.Resolve(&ResolveKey{
		Source:      msg.GetOperation().ConsumerId,
		Destination: msg.ServiceName,
		RpcMethod:   RPCCheck,
	})
	glog.V(2).Infof("Resolved: %d checkers %#v ==> %#v", len(checkers), *msg, checkers)
	ce := []*sc.CheckError{}
	for _, checker := range checkers {
		glog.V(1).Infof("Checking %s %s", checker.Kind, msg)
		ru, converted := checker.Params.(*RuntimeAdapterState)
		if !converted {
			glog.Warningf("%s was not converted", checker.Kind)
			continue
		}

		chk, err := c.FindChecker(checker.Kind, ru)
		if err != nil {
			glog.Warningf("%s Could not get checker %s", checker.Kind, err)
			continue
		}

		cer, er := chk.Check(msg)
		if er != nil {
			cer = &sc.CheckError{
				Code:   sc.CheckError_PERMISSION_DENIED,
				Detail: er.Error(),
			}
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

func (c *CheckerManager) ConfigChange(cfg *ServicesConfig) {
	glog.V(1).Infof("ConfigChanged", *cfg)
	c.cfg.Store(cfg)
}

func (c *CheckerManager) Checkers() []Checker {
	return []Checker{}
}
