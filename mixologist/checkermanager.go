package mixologist

import (
	"errors"
	"github.com/cloudendpoints/mixologist/mixologist/config"
	"github.com/golang/glog"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/net/context"
	sc "google/api/servicecontrol/v1"
	"reflect"
	"strings"
)

// DecodeError -- decoder specific error
// contains a slice of required fields that were missed
type DecodeError struct {
	err     error
	Missing []string
}

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

	return Validate([]string{}, value, &md)
}

// Validate -- validate the filled struct with "required" and other tags
func Validate(name []string, value reflect.Value, md *mapstructure.Metadata) *DecodeError {
	var missing []string
	for i := 0; i < value.NumField(); i++ {
		fld := value.Type().Field(i)
		tag, ok := fld.Tag.Lookup("yaml")
		fldArr := append(name, fld.Name)
		fldName := strings.Join(fldArr, ".")
		vfld := value.Field(i)
		if !ok {
			tag = string(fld.Tag)
		}
		if strings.Contains(tag, "required") {
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
	return errors.New("Adater of type '" + atype + "' is not available")
}

// NewCheckerManager -- given a registry and a config object return a CheckerManager
func NewCheckerManager(registry map[string]CheckerBuilder, c config.ServicesConfig) (*CheckerManager, []error) {
	var erra []error
	glog.Infof("Creating checker manager")
	glog.V(2).Infof("checker manager config: %#v", c)
	allsvc := c[config.EveryService]
	checkerImpls := make([]Checker, 0, len(allsvc.Ingress.Checkers))
	ctx := config.EveryService + ".Ingress.Checkers"
	for _, checkerCfg := range allsvc.Ingress.Checkers {
		checkerName := checkerCfg.Kind
		glog.Info("Creating checker ", checkerName)
		if cn, ok := registry[checkerName]; ok {
			ccfg := cn.ConfigStruct()
			err := Decode(checkerCfg.Params, ccfg)
			if err != nil {
				glog.Errorf("ERROR: Invalid Params for Adapter Type '%s' in %s: %s\ninput: %v\noutput: %#v", checkerName, ctx, err, checkerCfg.Params, ccfg)
				erra = append(erra, err)
				continue
			}
			if err := cn.ValidateConfig(ccfg); err != nil {
				erra = append(erra, err)
				glog.Error("validation failed: ", checkerName, " ", err, " ", ccfg)
				continue
			}
			if cc, err := cn.BuildChecker(ccfg); cc != nil {
				glog.Info("Built checker: ", checkerName, " ", cc, ccfg)
				checkerImpls = append(checkerImpls, cc)
			} else {
				erra = append(erra, err)
				glog.Error("Unable to build checker: ", checkerName, " ", err)
			}
		} else {
			ers := ErrAdapterUnavailable(checkerName)
			glog.Warning(ers)
			erra = append(erra, ers)
		}
	}
	glog.Info("Available Checkers: ", len(checkerImpls))
	return &CheckerManager{
		checkers: checkerImpls,
	}, erra
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

func (c *CheckerManager) Checkers() []Checker {
	return c.checkers
}
