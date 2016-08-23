package mixologist

import (
	"golang.org/x/net/context"
	sc "google/api/servicecontrol/v1"
)

type (
	// Server API for ServiceController service
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

	Handler struct {
		Server ServiceControllerServer
	}

	ControllerImpl struct {
	}
)
