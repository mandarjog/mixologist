package mixologist

import (
	"errors"
	"reflect"
	"strings"

	"github.com/golang/glog"
	"github.com/mitchellh/mapstructure"
)

const (
	// EveryService -- These rules apply to every service interaction
	EveryService = "_EVERY_SERVICE_"
)

type (
	// CacheParams -- configure caching behaviour
	// When adapter replies are cacheable, these params
	// configure behavious of the cache
	CacheParams struct {
	}

	// BatchParams -- configure batching behaviour of the adapter
	// These params should be used by the framework, *not* the adapter itself
	BatchParams struct {
		Size       int
		TimeoutSec int
	}
	// ConstructorParams -- 'Kind' is the adapter type
	// And Params are passed to the Kind constructor
	// This struct is sufficient to create a adapter
	ConstructorParams struct {
		// Kind - type of adapter
		Kind string
		// opaque params passed to the adapter
		Params interface{} `yaml:",omitempty"`
	}
	// AdapterParams -- constructor params and associated params
	AdapterParams struct {
		// Identifier of this adapter
		// should be unique within Kind, optional
		ID string
		// ConstructorParams embedded
		ConstructorParams `yaml:",omitempty,inline"`

		// cache params
		CacheParams CacheParams `yaml:",omitempty"`

		// batching params
		BatchParams BatchParams `yaml:",omitempty"`
	}

	// AdapterConfig -- in the given context
	// represents adapter configrations for all defined functions
	AdapterConfig struct {
		Checkers  []*AdapterParams `yaml:",omitempty"`
		Reporters []*AdapterParams `yaml:",omitempty"`
		// other functions
	}
	// BindingConfig -- same as adapter config + service
	// The other end of the service is understood by context
	BindingConfig struct {
		// ServiceID the service that this adapter config applies to
		ServiceID string
		// AdapterConfig -- embedded adapterconfig
		Adapters *AdapterConfig `yaml:",omitempty"`
	}

	// ServiceConfig -- Configuration from ServiceID's point of view
	ServiceConfig struct {
		// ServiceID the service that this adapter config applies to
		// this ServiceID also 'owns' the datastructure
		ServiceID string
		// Self -- configuration that applies to the service itself.
		// When service directly issues Check or Log / Report calls
		Self *AdapterConfig `yaml:",omitempty"`
		// Ingress - applies When the given service is 'Producer'
		// irrespective of the calling entity
		Ingress *AdapterConfig `yaml:",omitempty"`
		// Egress - applies When the given service is 'Consumer'
		// irrespective of the called entity
		Egress *AdapterConfig `yaml:",omitempty"`
		// Consumers -- Applies to binding between binding.ServiceID --> ServiceID
		// binding.ServiceID is the consumer service of 'ServiceID'
		Consumers map[string]*BindingConfig `yaml:",omitempty"`
		// Producers -- Applies to binding between ServiceID --> binding.ServiceID
		// binding.ServiceID is the Producer service for 'ServiceID'
		Producers map[string]*BindingConfig `yaml:",omitempty"`
	}

	// ServicesConfig -- toplevel map describing all known services
	ServicesConfig map[string]*ServiceConfig

	RPCMethod string
	// config resolver
	ResolveKey struct {
		Source      string
		Destination string
		RpcMethod   RPCMethod
	}

	// ConfigChanger -- called when a new config is available
	ConfigChanger interface {
		// ConfigChange -- update config
		ConfigChange(cfg *ServicesConfig)
	}
	// DecodeError -- decoder specific error
	// contains a slice of required fields that were missed
	DecodeError struct {
		err     error
		Missing []string
	}
)

const (
	RPCCheck  RPCMethod = "CHECK"
	RPCReport RPCMethod = "REPORT"
)

// Error -- conform to error interface
func (e DecodeError) Error() string {
	return e.err.Error()
}

// DE -- Create a new decoder error
func NewDecoderError(err error) *DecodeError {
	return &DecodeError{
		err: err,
	}
}

// Decode -- convert generic interface into the specific struct
// that was provided by the adapter
// If the struct is tagged with 'required' fields, appropriate errors
// are returned.
func Decode(src interface{}, dest interface{}) *DecodeError {
	var md mapstructure.Metadata
	mcfg := mapstructure.DecoderConfig{
		Metadata: &md,
		Result:   dest,
	}
	decoder, err := mapstructure.NewDecoder(&mcfg)
	if err != nil {
		return NewDecoderError(err)
	}
	err = decoder.Decode(src)
	if err != nil {
		return NewDecoderError(err)
	}
	// Check if any required keys are missing
	value := reflect.Indirect(reflect.ValueOf(dest))
	er := Validate([]string{}, value, &md)
	glog.V(2).Infof("Validating %#v, %#v ==> %#v", value, md, er)
	return er
}

// Validate -- validate the filled struct with "required" and other tags
func Validate(name []string, value reflect.Value, md *mapstructure.Metadata) *DecodeError {
	var missing []string
	for i := 0; i < value.NumField(); i++ {
		fld := value.Type().Field(i)
		tag, ok := fld.Tag.Lookup("required")
		fldArr := append(name, fld.Name)
		fldName := strings.Join(fldArr, ".")
		vfld := value.Field(i)
		if !ok {
			tag = string(fld.Tag)
		}
		if ok || strings.Contains(tag, "required") {
			found := false
			for _, k := range md.Keys {
				if k == fldName {
					found = true
					break
				}
			}
			if !found {
				glog.Errorf("%#v >> %#v >> not found %s, %#v", fld, vfld, fld.Name, md.Keys)
				missing = append(missing, fldName)
			}
		}
		if vfld.Kind() == reflect.Struct {
			er := Validate(fldArr, vfld, md)
			if er != nil {
				missing = append(missing, er.Missing...)
			}
		}
	}
	if len(missing) > 0 {
		return &DecodeError{
			err:     errors.New("Missing " + strings.Join(missing, ",")),
			Missing: missing,
		}
	}
	return nil
}

func ErrAdapterUnavailable(atype string) error {
	return errors.New("Adapter of type '" + atype + "' is not available")
}
