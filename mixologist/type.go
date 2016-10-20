package mixologist

import (
	sc "google/api/servicecontrol/v1"
	"io"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
)

type (

	// Config stores global config for all mixologist.
	Config struct {
		ReportConsumers  []string
		Checkers         []string
		Logging          LogsConfig
		WhiteListBackEnd string
	}

	LogsConfig struct {
		Backends   []string
		UseDefault bool
	}

	// ServiceControllerServer API
	// Same interface as used by grpc
	ServiceControllerServer interface {
		// Checks an operation with the Google Service Controller for features like
		// project status, billing status, abuse status, quota etc to decide whether
		// the given operation should proceed. It should be called before the
		// operation is executed.
		//
		// This method requires the `servicemanagement.services.check` permission
		// on the specified service. For more information, see
		// [Google Cloud IAM](https://cloud.google.com/iam).
		Check(context.Context, *sc.CheckRequest) (*sc.CheckResponse, error)
		// Reports operations to the Google Service Controller for features like
		// billing, logging, monitoring, quota usage, etc. It should be called
		// after the operation is completed.
		//
		// This method requires the `servicemanagement.services.report` permission
		// on the specified service. For more information, see
		// [Google Cloud IAM](https://cloud.google.com/iam).
		Report(context.Context, *sc.ReportRequest) (*sc.ReportResponse, error)
	}

	// Controller -- main ServerController interface augmented with Other fns
	Controller interface {
		ServiceControllerServer
		ReportQueue() chan *sc.ReportRequest
	}

	// Readfn -- used to read from http body. used for error injection
	readfn      func(r io.Reader) (msg []byte, err error)
	unmarshalfn func(buf []byte, pb proto.Message) error
	marshalfn   func(pb proto.Message) (buf []byte, err error)
	// Handler -- Main handler and the handler map
	Handler struct {
		Server         ServiceControllerServer
		ReportHandlers []*PrefixAndHandler

		// private type when alternate impl is provided at construction time
		readf     readfn
		marshal   marshalfn
		unmarshal unmarshalfn
	}

	CheckerManager struct {
		cfg atomic.Value

		lock     sync.RWMutex
		checkers map[ConstructorParams]Checker
	}
	// ControllerImpl -- The controller that is implemented by framework itself
	// It delelegates the actual work to a the *real* ServiceControllerServer
	ControllerImpl struct {
		reportQueue    chan *sc.ReportRequest
		checkerManager *CheckerManager
	}

	// ReportConsumerManagerImpl -- store consumer manager config/state
	ReportConsumerManagerImpl struct {
		reportQueue chan *sc.ReportRequest
		consumers   []ReportConsumer
	}
	// PrefixAndHandler -- as the name suggests, returned by consumers if they wish to have
	// a listener
	PrefixAndHandler struct {
		Prefix  string
		Handler http.Handler
	}

	// ReportConsumer -- components that wish to consume ReportRequest messages
	ReportConsumer interface {
		// Name -- name of this consumer
		//FIXME change to Name()
		GetName() string
		// Consume report
		Consume([]*sc.ReportRequest) error
		// Get path mapping and handler
		// can return nil
		//FIXME change to PrefixAndHandler
		GetPrefixAndHandler() *PrefixAndHandler
	}
	//ReportConsumerBuilder -- Every report consumer should register its builder
	// in the init method
	ReportConsumerBuilder interface {
		// Given an arbitrary map create a new consumer
		BuildConsumer(Config) (ReportConsumer, error)
	}

	Unloader interface {
		// Unload -- Called when this checker is no longer needed
		Unload()
	}

	// Checker -- components that wish to perform checks on a request
	Checker interface {
		// Name -- name of this checker
		Name() string
		// Check -- check if the current request should go thru
		// per this checker
		Check(*sc.CheckRequest) (*sc.CheckError, error)
		// Unload -- called when this adapter is no longer needed
		Unloader
	}

	AdapterBuilder interface {
		// ConfigStruct -- return a pointer to an instance of a struct needed to configure this checker
		// Typically the implementation is a one-liner. ex:
		// return &CheckerConfig{}
		ConfigStruct() (confPtr interface{})

		// ValidateConfig -- validate configuration struct. Return error if validation fails
		ValidateConfig(conf interface{}) error
	}

	// CheckerBuilder -- build an instance of a checker
	CheckerBuilder interface {
		// AdapterBuilder -- embedded
		AdapterBuilder
		// BuildChecker -- given a pointer to a properly filled struct obtained from ConfigStruct(),
		// return an initialized checker
		BuildChecker(interface{}) (Checker, error)
	}
)
