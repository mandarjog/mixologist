// Code generated by protoc-gen-go.
// source: google/api/servicecontrol/v1/check_error.proto
// DO NOT EDIT!

/*
Package servicecontrol is a generated protocol buffer package.

It is generated from these files:
	google/api/servicecontrol/v1/check_error.proto
	google/api/servicecontrol/v1/distribution.proto
	google/api/servicecontrol/v1/log_entry.proto
	google/api/servicecontrol/v1/metric_value.proto
	google/api/servicecontrol/v1/operation.proto
	google/api/servicecontrol/v1/quota_properties.proto
	google/api/servicecontrol/v1/service_controller.proto

It has these top-level messages:
	CheckError
	Distribution
	LogEntry
	MetricValue
	MetricValueSet
	Operation
	QuotaProperties
	CheckRequest
	CheckResponse
	ReportRequest
	ReportResponse
*/
package servicecontrol

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "google.golang.org/genproto/googleapis/api/serviceconfig"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Error codes for the check errors.
type CheckError_Code int32

const (
	// This is never used in `CheckResponse`.
	CheckError_ERROR_CODE_UNSPECIFIED CheckError_Code = 0
	// The consumer's project id is not found.
	// Same as [google.rpc.Code.NOT_FOUND][].
	CheckError_NOT_FOUND CheckError_Code = 5
	// The consumer doesn't have access to the specified resource.
	// Same as [google.rpc.Code.PERMISSION_DENIED][].
	CheckError_PERMISSION_DENIED CheckError_Code = 7
	// Quota check failed. Same as [google.rpc.Code.RESOURCE_EXHAUSTED][].
	CheckError_RESOURCE_EXHAUSTED CheckError_Code = 8
	// Budget check failed.
	CheckError_BUDGET_EXCEEDED CheckError_Code = 100
	// The request has been flagged as a DoS attack.
	CheckError_DENIAL_OF_SERVICE_DETECTED CheckError_Code = 101
	// The request should be rejected in order to protect the service from
	// being overloaded.
	CheckError_LOAD_SHEDDING CheckError_Code = 102
	// The consumer has been flagged as an abuser.
	CheckError_ABUSER_DETECTED CheckError_Code = 103
	// The consumer hasn't activated the service.
	CheckError_SERVICE_NOT_ACTIVATED CheckError_Code = 104
	// The consumer cannot access the service due to visibility configuration.
	CheckError_VISIBILITY_DENIED CheckError_Code = 106
	// The consumer cannot access the service because billing is disabled.
	CheckError_BILLING_DISABLED CheckError_Code = 107
	// Consumer's project has been marked as deleted (soft deletion).
	CheckError_PROJECT_DELETED CheckError_Code = 108
	// Consumer's project number or id does not represent a valid project.
	CheckError_PROJECT_INVALID CheckError_Code = 114
	// Consumer's project does not allow requests from this IP address.
	CheckError_IP_ADDRESS_BLOCKED CheckError_Code = 109
	// Consumer's project does not allow requests from this referer address.
	CheckError_REFERER_BLOCKED CheckError_Code = 110
	// Consumer's project does not allow requests from this client application.
	CheckError_CLIENT_APP_BLOCKED CheckError_Code = 111
	// The consumer's API key is invalid.
	CheckError_API_KEY_INVALID CheckError_Code = 105
	// Consumer's API Key has expired.
	CheckError_API_KEY_EXPIRED CheckError_Code = 112
	// Consumer's API Key not found in config record.
	CheckError_API_KEY_NOT_FOUND CheckError_Code = 113
	// Consumer's spatula header is invalid.
	CheckError_SPATULA_HEADER_INVALID CheckError_Code = 115
	// The backend server for looking up project id/number is unavailable.
	CheckError_NAMESPACE_LOOKUP_UNAVAILABLE CheckError_Code = 300
	// The backend server for checking service status is unavailable.
	CheckError_SERVICE_STATUS_UNAVAILABLE CheckError_Code = 301
	// The backend server for checking billing status is unavailable.
	CheckError_BILLING_STATUS_UNAVAILABLE CheckError_Code = 302
	// The quota check feature is temporarily unavailable:
	//  Could be due to either internal config error or server error
	CheckError_QUOTA_CHECK_UNAVAILABLE CheckError_Code = 303
)

var CheckError_Code_name = map[int32]string{
	0:   "ERROR_CODE_UNSPECIFIED",
	5:   "NOT_FOUND",
	7:   "PERMISSION_DENIED",
	8:   "RESOURCE_EXHAUSTED",
	100: "BUDGET_EXCEEDED",
	101: "DENIAL_OF_SERVICE_DETECTED",
	102: "LOAD_SHEDDING",
	103: "ABUSER_DETECTED",
	104: "SERVICE_NOT_ACTIVATED",
	106: "VISIBILITY_DENIED",
	107: "BILLING_DISABLED",
	108: "PROJECT_DELETED",
	114: "PROJECT_INVALID",
	109: "IP_ADDRESS_BLOCKED",
	110: "REFERER_BLOCKED",
	111: "CLIENT_APP_BLOCKED",
	105: "API_KEY_INVALID",
	112: "API_KEY_EXPIRED",
	113: "API_KEY_NOT_FOUND",
	115: "SPATULA_HEADER_INVALID",
	300: "NAMESPACE_LOOKUP_UNAVAILABLE",
	301: "SERVICE_STATUS_UNAVAILABLE",
	302: "BILLING_STATUS_UNAVAILABLE",
	303: "QUOTA_CHECK_UNAVAILABLE",
}
var CheckError_Code_value = map[string]int32{
	"ERROR_CODE_UNSPECIFIED":       0,
	"NOT_FOUND":                    5,
	"PERMISSION_DENIED":            7,
	"RESOURCE_EXHAUSTED":           8,
	"BUDGET_EXCEEDED":              100,
	"DENIAL_OF_SERVICE_DETECTED":   101,
	"LOAD_SHEDDING":                102,
	"ABUSER_DETECTED":              103,
	"SERVICE_NOT_ACTIVATED":        104,
	"VISIBILITY_DENIED":            106,
	"BILLING_DISABLED":             107,
	"PROJECT_DELETED":              108,
	"PROJECT_INVALID":              114,
	"IP_ADDRESS_BLOCKED":           109,
	"REFERER_BLOCKED":              110,
	"CLIENT_APP_BLOCKED":           111,
	"API_KEY_INVALID":              105,
	"API_KEY_EXPIRED":              112,
	"API_KEY_NOT_FOUND":            113,
	"SPATULA_HEADER_INVALID":       115,
	"NAMESPACE_LOOKUP_UNAVAILABLE": 300,
	"SERVICE_STATUS_UNAVAILABLE":   301,
	"BILLING_STATUS_UNAVAILABLE":   302,
	"QUOTA_CHECK_UNAVAILABLE":      303,
}

func (x CheckError_Code) String() string {
	return proto.EnumName(CheckError_Code_name, int32(x))
}
func (CheckError_Code) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

// Defines the errors to be returned in
// [google.api.servicecontrol.v1.CheckResponse.check_errors][google.api.servicecontrol.v1.CheckResponse.check_errors].
type CheckError struct {
	// The error code.
	Code CheckError_Code `protobuf:"varint,1,opt,name=code,enum=google.api.servicecontrol.v1.CheckError_Code" json:"code,omitempty"`
	// The error detail.
	//
	Detail string `protobuf:"bytes,2,opt,name=detail" json:"detail,omitempty"`
}

func (m *CheckError) Reset()                    { *m = CheckError{} }
func (m *CheckError) String() string            { return proto.CompactTextString(m) }
func (*CheckError) ProtoMessage()               {}
func (*CheckError) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func init() {
	proto.RegisterType((*CheckError)(nil), "google.api.servicecontrol.v1.CheckError")
	proto.RegisterEnum("google.api.servicecontrol.v1.CheckError_Code", CheckError_Code_name, CheckError_Code_value)
}

func init() { proto.RegisterFile("google/api/servicecontrol/v1/check_error.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 574 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x7c, 0x93, 0xdf, 0x52, 0x13, 0x31,
	0x14, 0xc6, 0x6d, 0x07, 0x50, 0x32, 0x83, 0x2c, 0x51, 0x10, 0x3b, 0x8c, 0x22, 0x57, 0xdc, 0xb8,
	0x1d, 0xf4, 0xc6, 0xdb, 0x6c, 0x72, 0x4a, 0x63, 0xc3, 0x26, 0x26, 0xbb, 0x1d, 0xb8, 0xca, 0xac,
	0xed, 0xb2, 0xac, 0x94, 0x4d, 0xdd, 0x76, 0x78, 0x2a, 0xff, 0xbc, 0x81, 0x4f, 0xe1, 0xc3, 0x78,
	0xe9, 0xa4, 0xd0, 0x16, 0x1c, 0x87, 0xcb, 0xfd, 0xce, 0xef, 0xfb, 0xce, 0x39, 0x49, 0x16, 0x85,
	0x85, 0x73, 0xc5, 0x28, 0x6f, 0x67, 0xe3, 0xb2, 0x3d, 0xc9, 0xeb, 0xeb, 0x72, 0x90, 0x0f, 0x5c,
	0x35, 0xad, 0xdd, 0xa8, 0x7d, 0x7d, 0xd4, 0x1e, 0x5c, 0xe4, 0x83, 0x4b, 0x9b, 0xd7, 0xb5, 0xab,
	0xc3, 0x71, 0xed, 0xa6, 0x0e, 0xef, 0xdd, 0xf0, 0x61, 0x36, 0x2e, 0xc3, 0xfb, 0x7c, 0x78, 0x7d,
	0xd4, 0xe2, 0xb7, 0xd5, 0xc2, 0x8d, 0xb2, 0xaa, 0x08, 0x5d, 0x5d, 0xb4, 0x8b, 0xbc, 0x9a, 0x39,
	0xdb, 0x37, 0xa5, 0x6c, 0x5c, 0x4e, 0xfe, 0x69, 0x76, 0x5e, 0x16, 0xed, 0xac, 0xaa, 0xdc, 0x34,
	0x9b, 0x96, 0xae, 0x9a, 0xdc, 0x34, 0x3a, 0xf8, 0xb5, 0x8a, 0x10, 0xf5, 0xed, 0xc1, 0x77, 0xc7,
	0x04, 0xad, 0x0c, 0xdc, 0x30, 0xdf, 0x6d, 0xec, 0x37, 0x0e, 0x9f, 0xbe, 0x7b, 0x1b, 0x3e, 0x34,
	0x46, 0xb8, 0xf4, 0x85, 0xd4, 0x0d, 0x73, 0x3d, 0xb3, 0xe2, 0x1d, 0xb4, 0x36, 0xcc, 0xa7, 0x59,
	0x39, 0xda, 0x6d, 0xee, 0x37, 0x0e, 0xd7, 0xf5, 0xed, 0xd7, 0xc1, 0xef, 0x15, 0xb4, 0xe2, 0x31,
	0xdc, 0x42, 0x3b, 0xa0, 0xb5, 0xd4, 0x96, 0x4a, 0x06, 0x36, 0x8d, 0x8d, 0x02, 0xca, 0x3b, 0x1c,
	0x58, 0xf0, 0x08, 0x6f, 0xa0, 0xf5, 0x58, 0x26, 0xb6, 0x23, 0xd3, 0x98, 0x05, 0xab, 0x78, 0x1b,
	0x6d, 0x29, 0xd0, 0x27, 0xdc, 0x18, 0x2e, 0x63, 0xcb, 0x20, 0xf6, 0xd4, 0x63, 0xbc, 0x83, 0xb0,
	0x06, 0x23, 0x53, 0x4d, 0xc1, 0xc2, 0x69, 0x97, 0xa4, 0x26, 0x01, 0x16, 0x3c, 0xc1, 0xcf, 0xd0,
	0x66, 0x94, 0xb2, 0x63, 0x48, 0x2c, 0x9c, 0x52, 0x00, 0x06, 0x2c, 0x18, 0xe2, 0x57, 0xa8, 0xe5,
	0x8d, 0x44, 0x58, 0xd9, 0xb1, 0x06, 0x74, 0x9f, 0x53, 0xb0, 0x0c, 0x12, 0xa0, 0xde, 0x94, 0xe3,
	0x2d, 0xb4, 0x21, 0x24, 0x61, 0xd6, 0x74, 0x81, 0x31, 0x1e, 0x1f, 0x07, 0xe7, 0x3e, 0x87, 0x44,
	0xa9, 0x01, 0xbd, 0xe4, 0x0a, 0xfc, 0x12, 0x6d, 0xcf, 0xdd, 0x7e, 0x44, 0x42, 0x13, 0xde, 0x27,
	0xbe, 0x74, 0xe1, 0xc7, 0xec, 0x73, 0xc3, 0x23, 0x2e, 0x78, 0x72, 0x36, 0x1f, 0xf3, 0x0b, 0x7e,
	0x8e, 0x82, 0x88, 0x0b, 0xc1, 0xe3, 0x63, 0xcb, 0xb8, 0x21, 0x91, 0x00, 0x16, 0x5c, 0xfa, 0x70,
	0xa5, 0xe5, 0x47, 0xa0, 0x89, 0x65, 0x20, 0xc0, 0x27, 0x8c, 0xee, 0x8a, 0x3c, 0xee, 0x13, 0xc1,
	0x59, 0x50, 0xfb, 0x35, 0xb9, 0xb2, 0x84, 0x31, 0x0d, 0xc6, 0xd8, 0x48, 0x48, 0xda, 0x03, 0x16,
	0x5c, 0x79, 0x58, 0x43, 0x07, 0x34, 0xe8, 0x85, 0x58, 0x79, 0x98, 0x0a, 0x0e, 0x71, 0x62, 0x89,
	0x52, 0x0b, 0xdd, 0xcd, 0x76, 0x51, 0xdc, 0xf6, 0xe0, 0x6c, 0x91, 0x5c, 0xde, 0x15, 0xe1, 0x54,
	0x71, 0x0d, 0x2c, 0x18, 0xfb, 0x2d, 0xe6, 0xe2, 0xf2, 0x0e, 0xbe, 0xfa, 0xeb, 0x32, 0x8a, 0x24,
	0xa9, 0x20, 0xb6, 0x0b, 0x84, 0x81, 0x5e, 0xe4, 0x4c, 0xf0, 0x1b, 0xb4, 0x17, 0x93, 0x13, 0x30,
	0x8a, 0x50, 0xb0, 0x42, 0xca, 0x5e, 0xaa, 0x6c, 0x1a, 0x93, 0x3e, 0xe1, 0xc2, 0xaf, 0x1b, 0x7c,
	0x6b, 0xe2, 0xd7, 0xa8, 0x35, 0x3f, 0x36, 0x93, 0x90, 0x24, 0x35, 0xf7, 0x80, 0xef, 0x33, 0x60,
	0x7e, 0x4a, 0xff, 0x01, 0x7e, 0x34, 0xf1, 0x1e, 0x7a, 0xf1, 0x29, 0x95, 0x09, 0xb1, 0xb4, 0x0b,
	0xb4, 0x77, 0xaf, 0xfa, 0xb3, 0x19, 0x7d, 0x40, 0xfb, 0x03, 0x77, 0xf5, 0xe0, 0x43, 0x8d, 0x36,
	0x97, 0x2f, 0x55, 0xf9, 0x57, 0xaf, 0x1a, 0x7f, 0x1a, 0x8d, 0xcf, 0x6b, 0xb3, 0x3f, 0xe0, 0xfd,
	0xdf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x46, 0x54, 0xf0, 0x6b, 0x9c, 0x03, 0x00, 0x00,
}
