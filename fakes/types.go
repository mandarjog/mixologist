package fakes

import (
	sc "google/api/servicecontrol/v1"
)

type (
	consumer struct {
		name string
		meta map[string]interface{}
		msgs []*sc.ReportRequest
	}
	builder struct {
		name string
	}
)
