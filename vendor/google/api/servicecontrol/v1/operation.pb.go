// Code generated by protoc-gen-go.
// source: google/api/servicecontrol/v1/operation.proto
// DO NOT EDIT!

package servicecontrol

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "google.golang.org/genproto/googleapis/api/serviceconfig"
import google_protobuf3 "github.com/golang/protobuf/ptypes/timestamp"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Defines the importance of the data contained in the operation.
type Operation_Importance int32

const (
	// The operation doesn't contain significant monetary value or audit
	// trail. The API implementation may cache and aggregate the data.
	// There is no deduplication based on `operation_id`. The data
	// may be lost when rare and unexpected system failures occur.
	Operation_LOW Operation_Importance = 0
	// The operation contains significant monetary value or audit trail.
	// The API implementation doesn't cache and aggregate the data.
	// Deduplication based on `operation_id` is performed for monetary
	// values. If the method returns successfully, it's guaranteed that
	// the data are persisted in durable storage.
	Operation_HIGH Operation_Importance = 1
)

var Operation_Importance_name = map[int32]string{
	0: "LOW",
	1: "HIGH",
}
var Operation_Importance_value = map[string]int32{
	"LOW":  0,
	"HIGH": 1,
}

func (x Operation_Importance) String() string {
	return proto.EnumName(Operation_Importance_name, int32(x))
}
func (Operation_Importance) EnumDescriptor() ([]byte, []int) { return fileDescriptor4, []int{0, 0} }

// Represents information regarding an operation.
type Operation struct {
	// Identity of the operation. It must be unique within the scope of the
	// service that the operation is generated. If the service calls
	// Check() and Report() on the same operation, the two calls should carry
	// the same id.
	//
	// UUID version 4 is recommended, though not required.
	// In the scenarios where an operation is computed from existing information
	// and an idempotent id is desirable for deduplication purpose, UUID version 5
	// is recommended. See RFC 4122 for details.
	OperationId string `protobuf:"bytes,1,opt,name=operation_id,json=operationId" json:"operation_id,omitempty"`
	// Fully qualified name of the operation.
	// Example of an RPC method name used as operation name:
	//     google.example.library.v1.LibraryService.CreateShelf
	// Example of a service defined operation name:
	//     compute.googleapis.com/InstanceHeartbeat
	OperationName string `protobuf:"bytes,2,opt,name=operation_name,json=operationName" json:"operation_name,omitempty"`
	// Identity of the consumer who is using the service.
	// This field should be filled in for the operations initiated by a
	// consumer, but not for service initiated operations that are
	// not related to a specific consumer.
	//
	// The accepted format is dependent on the implementation.
	// The Google Service Control implementation accepts four forms:
	// "project:<project_id>", "project_number:<project_number>",
	// "api_key:<api_key>" and "spatula_header:<spatula_header>".
	ConsumerId string `protobuf:"bytes,3,opt,name=consumer_id,json=consumerId" json:"consumer_id,omitempty"`
	// Start time of the operation.
	// Required.
	StartTime *google_protobuf3.Timestamp `protobuf:"bytes,4,opt,name=start_time,json=startTime" json:"start_time,omitempty"`
	// End time of the operation.
	// Required when the operation is used in ControllerService.Report,
	// but optional when the operation is used in ControllerService.Check.
	EndTime *google_protobuf3.Timestamp `protobuf:"bytes,5,opt,name=end_time,json=endTime" json:"end_time,omitempty"`
	// Labels describing the operation. Only the following labels are allowed:
	//
	// Labels describing the monitored resource. The labels must be defined in
	// the service configuration.
	//
	// Default labels of the metric values. When specified, labels defined in the
	// metric value overrule.
	//
	// Labels are defined and documented by Google Cloud Platform. For example:
	// `cloud.googleapis.com/location: "us-east1"`.
	Labels map[string]string `protobuf:"bytes,6,rep,name=labels" json:"labels,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	// Represents information about this operation. Each MetricValueSet
	// corresponds to a metric defined in the service configuration.
	// The data type used in the MetricValueSet must agree with
	// the data type specified in the metric definition.
	//
	// Within a single operation, it is not allowed to have more than one
	// MetricValue instances that have the same metric names and identical
	// label value combinations. The existence of such duplicated MetricValue
	// instances in a request causes the entire request being rejected with
	// an invalid argument error.
	MetricValueSets []*MetricValueSet `protobuf:"bytes,7,rep,name=metric_value_sets,json=metricValueSets" json:"metric_value_sets,omitempty"`
	// Represents information to be logged.
	LogEntries []*LogEntry `protobuf:"bytes,8,rep,name=log_entries,json=logEntries" json:"log_entries,omitempty"`
	// Represents the properties needed for quota check. Applicable only if this
	// operation is for a quota check request.
	QuotaProperties *QuotaProperties `protobuf:"bytes,9,opt,name=quota_properties,json=quotaProperties" json:"quota_properties,omitempty"`
	// The importance of the data contained in the operation.
	Importance Operation_Importance `protobuf:"varint,11,opt,name=importance,enum=google.api.servicecontrol.v1.Operation_Importance" json:"importance,omitempty"`
}

func (m *Operation) Reset()                    { *m = Operation{} }
func (m *Operation) String() string            { return proto.CompactTextString(m) }
func (*Operation) ProtoMessage()               {}
func (*Operation) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{0} }

func (m *Operation) GetStartTime() *google_protobuf3.Timestamp {
	if m != nil {
		return m.StartTime
	}
	return nil
}

func (m *Operation) GetEndTime() *google_protobuf3.Timestamp {
	if m != nil {
		return m.EndTime
	}
	return nil
}

func (m *Operation) GetLabels() map[string]string {
	if m != nil {
		return m.Labels
	}
	return nil
}

func (m *Operation) GetMetricValueSets() []*MetricValueSet {
	if m != nil {
		return m.MetricValueSets
	}
	return nil
}

func (m *Operation) GetLogEntries() []*LogEntry {
	if m != nil {
		return m.LogEntries
	}
	return nil
}

func (m *Operation) GetQuotaProperties() *QuotaProperties {
	if m != nil {
		return m.QuotaProperties
	}
	return nil
}

func init() {
	proto.RegisterType((*Operation)(nil), "google.api.servicecontrol.v1.Operation")
	proto.RegisterEnum("google.api.servicecontrol.v1.Operation_Importance", Operation_Importance_name, Operation_Importance_value)
}

func init() { proto.RegisterFile("google/api/servicecontrol/v1/operation.proto", fileDescriptor4) }

var fileDescriptor4 = []byte{
	// 529 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x8c, 0x52, 0xcf, 0x6f, 0xd3, 0x30,
	0x18, 0x25, 0xeb, 0x7e, 0xb4, 0x5f, 0xa0, 0x2b, 0x16, 0x87, 0xa8, 0x42, 0x5a, 0x99, 0x04, 0xea,
	0x61, 0xc4, 0x5a, 0x2b, 0x04, 0x83, 0xdb, 0x24, 0xb4, 0x55, 0x14, 0x56, 0x02, 0x02, 0x6e, 0x91,
	0x9b, 0x7e, 0x0b, 0x16, 0x89, 0x9d, 0xd9, 0x6e, 0xa5, 0xfe, 0xbb, 0xfc, 0x15, 0x1c, 0x51, 0x9c,
	0x1f, 0x6d, 0x77, 0x08, 0xdc, 0xec, 0xcf, 0xef, 0x3d, 0x3f, 0x3f, 0x3f, 0x38, 0x8b, 0xa5, 0x8c,
	0x13, 0xa4, 0x2c, 0xe3, 0x54, 0xa3, 0x5a, 0xf1, 0x08, 0x23, 0x29, 0x8c, 0x92, 0x09, 0x5d, 0x9d,
	0x53, 0x99, 0xa1, 0x62, 0x86, 0x4b, 0xe1, 0x67, 0x4a, 0x1a, 0x49, 0x9e, 0x16, 0x68, 0x9f, 0x65,
	0xdc, 0xdf, 0x45, 0xfb, 0xab, 0xf3, 0xfe, 0xa4, 0x3c, 0x8d, 0x65, 0xc2, 0x44, 0xec, 0x4b, 0x15,
	0xd3, 0x18, 0x85, 0x65, 0xd2, 0xe2, 0x88, 0x65, 0x5c, 0xdf, 0xbb, 0xea, 0x96, 0xc7, 0x94, 0x09,
	0x21, 0x8d, 0xbd, 0x47, 0x17, 0x17, 0xf5, 0x9b, 0x6d, 0x25, 0x32, 0x0e, 0x51, 0x18, 0xb5, 0x2e,
	0xd1, 0xb4, 0x11, 0x9d, 0xa2, 0x51, 0x3c, 0x0a, 0x57, 0x2c, 0x59, 0x62, 0x49, 0x18, 0x37, 0x12,
	0xee, 0x96, 0xd2, 0xb0, 0x30, 0x53, 0xf9, 0xeb, 0x0d, 0xc7, 0xca, 0xd3, 0xbb, 0x98, 0x9b, 0x9f,
	0xcb, 0xb9, 0x1f, 0xc9, 0x94, 0x16, 0x4f, 0xa4, 0xf6, 0x60, 0xbe, 0xbc, 0xa5, 0x99, 0x59, 0x67,
	0xa8, 0xa9, 0xe1, 0x29, 0x6a, 0xc3, 0xd2, 0x6c, 0xb3, 0x2a, 0xc8, 0xa7, 0xbf, 0x0f, 0xa0, 0x73,
	0x53, 0xa5, 0x49, 0x9e, 0xc1, 0xc3, 0x3a, 0xda, 0x90, 0x2f, 0x3c, 0x67, 0xe0, 0x0c, 0x3b, 0x81,
	0x5b, 0xcf, 0x26, 0x0b, 0xf2, 0x1c, 0xba, 0x1b, 0x88, 0x60, 0x29, 0x7a, 0x7b, 0x16, 0xf4, 0xa8,
	0x9e, 0x7e, 0x62, 0x29, 0x92, 0x13, 0x70, 0x23, 0x29, 0xf4, 0x32, 0x45, 0x95, 0x0b, 0xb5, 0x2c,
	0x06, 0xaa, 0xd1, 0x64, 0x41, 0x2e, 0x00, 0xb4, 0x61, 0xca, 0x84, 0xb9, 0x23, 0x6f, 0x7f, 0xe0,
	0x0c, 0xdd, 0x51, 0xdf, 0x2f, 0x7f, 0xaa, 0xf2, 0xef, 0x7f, 0xad, 0xec, 0x06, 0x1d, 0x8b, 0xce,
	0xf7, 0xe4, 0x15, 0xb4, 0x51, 0x2c, 0x0a, 0xe2, 0xc1, 0x3f, 0x89, 0x47, 0x28, 0x16, 0x96, 0xf6,
	0x01, 0x0e, 0x13, 0x36, 0xc7, 0x44, 0x7b, 0x87, 0x83, 0xd6, 0xd0, 0x1d, 0x8d, 0xfd, 0xa6, 0xd6,
	0xf8, 0x75, 0x2a, 0xfe, 0xd4, 0xb2, 0xde, 0xe7, 0x1f, 0x1b, 0x94, 0x12, 0xe4, 0x07, 0x3c, 0xde,
	0xfe, 0xbf, 0x50, 0xa3, 0xd1, 0xde, 0x91, 0xd5, 0x3d, 0x6b, 0xd6, 0xfd, 0x68, 0x69, 0xdf, 0x72,
	0xd6, 0x17, 0x34, 0xc1, 0x71, 0xba, 0xb3, 0xd7, 0xe4, 0x0a, 0xdc, 0xaa, 0x47, 0x1c, 0xb5, 0xd7,
	0xb6, 0x9a, 0x2f, 0x9a, 0x35, 0xa7, 0x32, 0x2e, 0xec, 0x41, 0x52, 0xac, 0x38, 0xe6, 0x16, 0x7b,
	0xf7, 0x1b, 0xe3, 0x75, 0x6c, 0x5c, 0x2f, 0x9b, 0xd5, 0x3e, 0xe7, 0xac, 0x59, 0x4d, 0x0a, 0x8e,
	0xef, 0x76, 0x07, 0x24, 0x00, 0xe0, 0x69, 0x26, 0x95, 0x61, 0x22, 0x42, 0xcf, 0x1d, 0x38, 0xc3,
	0xee, 0x68, 0xf4, 0xbf, 0x69, 0x4e, 0x6a, 0x66, 0xb0, 0xa5, 0xd2, 0xbf, 0x00, 0x77, 0x2b, 0x67,
	0xd2, 0x83, 0xd6, 0x2f, 0x5c, 0x97, 0x05, 0xcc, 0x97, 0xe4, 0x09, 0x1c, 0xd8, 0xa8, 0xcb, 0xbe,
	0x15, 0x9b, 0xb7, 0x7b, 0x6f, 0x9c, 0xd3, 0x13, 0x80, 0x8d, 0x28, 0x39, 0x82, 0xd6, 0xf4, 0xe6,
	0x7b, 0xef, 0x01, 0x69, 0xc3, 0xfe, 0xf5, 0xe4, 0xea, 0xba, 0xe7, 0x5c, 0xbe, 0x86, 0x41, 0x24,
	0xd3, 0x46, 0x83, 0x97, 0xdd, 0xda, 0xe1, 0x2c, 0xef, 0xd0, 0xcc, 0xf9, 0xe3, 0x38, 0xf3, 0x43,
	0xdb, 0xa7, 0xf1, 0xdf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x55, 0x35, 0x49, 0x95, 0x8e, 0x04, 0x00,
	0x00,
}
