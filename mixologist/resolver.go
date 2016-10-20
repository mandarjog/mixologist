package mixologist

import (
	"fmt"

	"github.com/golang/glog"
)

type (
	RuntimeAdapterState struct {
		// Path to this node
		Path string
		// error encountered during conversion if any
		ConvertionError error
		TypedParams     interface{}
		Params          interface{}
		Builder         interface{}
	}
)

// Resolve -- Given a services config resolve it to an array
// of Adapters that should be dispatched.
// Note: config (cfg) is readonly -- so no locking is needed
// 	when Resolve runs concurrently
// TODO Add treatment of AdapterParams which includes caching and batching
func (cfg ServicesConfig) Resolve(msg *ResolveKey) (cp []*ConstructorParams) {
	if all, found := cfg[EveryService]; found {
		cp = append(cp, constructorParams(msg, all.Ingress, all.Egress, all.Self)...)
	}
	if src, found := cfg[msg.Source]; found {
		cp = append(cp, constructorParams(msg, src.Egress)...)
	}
	if dest, found := cfg[msg.Destination]; found {
		cp = append(cp, constructorParams(msg, dest.Ingress)...)
		if bnd, found := dest.Consumers[msg.Source]; found {
			cp = append(cp, constructorParams(msg, bnd.Adapters)...)
		}
	}
	glog.V(2).Infof("Resolved: %#v ==> %#v", *msg, len(cp))
	return cp
}

func adapterParams(ac *AdapterConfig, msg *ResolveKey) []*AdapterParams {
	switch msg.RpcMethod {
	case RPCCheck:
		return ac.Checkers
	case RPCReport:
		return ac.Reporters
	}
	return nil
}

// constructorParams -- Filter adapterconfig and return valid constroctorParams
func constructorParams(msg *ResolveKey, acs ...*AdapterConfig) []*ConstructorParams {
	cp := []*ConstructorParams{}
	for _, ac := range acs {
		if ac == nil {
			continue
		}
		for _, cc := range adapterParams(ac, msg) {
			ru, converted := cc.Params.(*RuntimeAdapterState)
			if !converted {
				glog.V(2).Infof("%s was not converted", cc.Kind)
				continue
			}
			if ru.ConvertionError != nil {
				glog.V(2).Infof("%s had conversion errors %s", cc.Kind, ru.ConvertionError)
				continue
			}
			cp = append(cp, &(cc.ConstructorParams))
		}
	}
	return cp
}

// ConvertParams -- traverses ServicesConfig and updates
// UnTyped ConstructorParams.Params to typed versions
func ConvertParams(cfg ServicesConfig, creg map[string]CheckerBuilder) (ServicesConfig, []error) {
	var erra []error
	for svcname, c := range cfg {
		erra = append(erra, updateAdapterConfig(svcname, creg, c.Egress, c.Ingress, c.Self)...)
		for bndname, bnd := range c.Consumers {
			erra = append(erra, updateAdapterConfig(svcname+".Consumers."+bndname, creg, bnd.Adapters)...)
		}
		for bndname, bnd := range c.Producers {
			erra = append(erra, updateAdapterConfig(svcname+".Producers."+bndname, creg, bnd.Adapters)...)
		}
	}
	return cfg, erra
}

func updateAdapterConfig(name string, creg map[string]CheckerBuilder, ac ...*AdapterConfig) []error {
	var erra []error
	for idx := range ac {
		if ac[idx] == nil {
			continue
		}
		erra = append(erra, updateAdapterParams(name+fmt.Sprintf("%s.%d", name, idx), creg, &(ac[idx].Checkers))...)
	}
	return erra
}

func updateAdapterParams(name string, reg map[string]CheckerBuilder, app *[]*AdapterParams) []error {
	var erra []error
	var badidx []int
	ap := *app
	for idx := range ap {
		if cn, ok := reg[ap[idx].Kind]; ok {
			ccfg := cn.ConfigStruct()
			ru, converted := ap[idx].Params.(*RuntimeAdapterState)
			if !converted {
				ru = &RuntimeAdapterState{
					Params:  ap[idx].Params,
					Builder: cn,
					Path:    name,
				}
				ap[idx].Params = ru
			}
			if err := Decode(ru.Params, ccfg); err != nil {
				erra = append(erra, err)
				ru.ConvertionError = err
				glog.Errorf("ERROR: Invalid Params for Adapter Type '%s' in %s: %s\ninput: %v\noutput: %#v", ap[idx].Kind, name, err, *ru, ccfg)
				continue
			}
			if err := cn.ValidateConfig(ccfg); err != nil {
				erra = append(erra, err)
				ru.ConvertionError = err
				glog.Errorf("ERROR: Invalid Params for Adapter Type '%s' in %s: %s\ninput: %v\noutput: %#v", ap[idx].Kind, name, err, *ru, ccfg)
				continue
			}
			ru.TypedParams = ccfg
		} else {
			badidx = append(badidx, idx)
			glog.Warningf("Unknown adapter type %s", ap[idx].Kind)
			erra = append(erra, ErrAdapterUnavailable(ap[idx].Kind))
		}
	}
	// remove bad idx from slice
	for i := len(badidx) - 1; i >= 0; i-- {
		idx := badidx[i]
		// remove idx
		ap[idx] = ap[len(ap)-1]
		ap[len(ap)-1] = nil
		ap = ap[:len(ap)-1]
	}
	*app = ap
	return erra
}
