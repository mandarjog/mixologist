package mixologist

import (
	"golang.org/x/net/context"
	sc "google/api/servicecontrol/v1"
	"net/http"
)

type (
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

	// Handler -- Main handler and the handler map
	Handler struct {
		Server         ServiceControllerServer
		ReportHandlers []*PrefixAndHandler
	}
	// ControllerImpl -- The controller that is implemented by framework itself
	// It delelegates the actual work to a the *real* ServiceControllerServer
	ControllerImpl struct {
		ReportQueue chan *sc.ReportRequest
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
		// GetName -- name of this consumer
		GetName() string
		// Consume report
		Consume(*sc.ReportRequest) error
		// Get path mapping and handler
		// can return nil
		GetPrefixAndHandler() *PrefixAndHandler
	}

	//ReportConsumerBuilder -- Every report consumer should register its builder
	// in the init method
	ReportConsumerBuilder interface {
		// Given an arbitrary map create a new consumer
		NewConsumer(map[string]interface{}) ReportConsumer
	}
)
