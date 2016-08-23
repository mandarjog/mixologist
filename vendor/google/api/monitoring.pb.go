// Code generated by protoc-gen-go.
// source: google/api/monitoring.proto
// DO NOT EDIT!

package api

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Monitoring configuration of the service.
//
// The example below shows how to configure monitored resources and metrics
// for monitoring. In the example, a monitored resource and two metrics are
// defined. The `library.googleapis.com/book/returned_count` metric is sent
// to both producer and consumer projects, whereas the
// `library.googleapis.com/book/overdue_count` metric is only sent to the
// consumer project.
//
//     monitored_resources:
//     - type: library.googleapis.com/branch
//       labels:
//       - key: /city
//         description: The city where the library branch is located in.
//       - key: /name
//         description: The name of the branch.
//     metrics:
//     - name: library.googleapis.com/book/returned_count
//       metric_kind: DELTA
//       value_type: INT64
//       labels:
//       - key: /customer_id
//     - name: library.googleapis.com/book/overdue_count
//       metric_kind: GAUGE
//       value_type: INT64
//       labels:
//       - key: /customer_id
//     monitoring:
//       producer_destinations:
//       - monitored_resource: library.googleapis.com/branch
//         metrics:
//         - library.googleapis.com/book/returned_count
//       consumer_destinations:
//       - monitored_resource: library.googleapis.com/branch
//         metrics:
//         - library.googleapis.com/book/returned_count
//         - library.googleapis.com/book/overdue_count
//
type Monitoring struct {
	// Monitoring configurations for sending metrics to the producer project.
	// There can be multiple producer destinations, each one must have a
	// different monitored resource type. A metric can be used in at most
	// one producer destination.
	ProducerDestinations []*Monitoring_MonitoringDestination `protobuf:"bytes,1,rep,name=producer_destinations,json=producerDestinations" json:"producer_destinations,omitempty"`
	// Monitoring configurations for sending metrics to the consumer project.
	// There can be multiple consumer destinations, each one must have a
	// different monitored resource type. A metric can be used in at most
	// one consumer destination.
	ConsumerDestinations []*Monitoring_MonitoringDestination `protobuf:"bytes,2,rep,name=consumer_destinations,json=consumerDestinations" json:"consumer_destinations,omitempty"`
}

func (m *Monitoring) Reset()                    { *m = Monitoring{} }
func (m *Monitoring) String() string            { return proto.CompactTextString(m) }
func (*Monitoring) ProtoMessage()               {}
func (*Monitoring) Descriptor() ([]byte, []int) { return fileDescriptor14, []int{0} }

func (m *Monitoring) GetProducerDestinations() []*Monitoring_MonitoringDestination {
	if m != nil {
		return m.ProducerDestinations
	}
	return nil
}

func (m *Monitoring) GetConsumerDestinations() []*Monitoring_MonitoringDestination {
	if m != nil {
		return m.ConsumerDestinations
	}
	return nil
}

// Configuration of a specific monitoring destination (the producer project
// or the consumer project).
type Monitoring_MonitoringDestination struct {
	// The monitored resource type. The type must be defined in
	// [Service.monitored_resources][google.api.Service.monitored_resources] section.
	MonitoredResource string `protobuf:"bytes,1,opt,name=monitored_resource,json=monitoredResource" json:"monitored_resource,omitempty"`
	// Names of the metrics to report to this monitoring destination.
	// Each name must be defined in [Service.metrics][google.api.Service.metrics] section.
	Metrics []string `protobuf:"bytes,2,rep,name=metrics" json:"metrics,omitempty"`
}

func (m *Monitoring_MonitoringDestination) Reset()         { *m = Monitoring_MonitoringDestination{} }
func (m *Monitoring_MonitoringDestination) String() string { return proto.CompactTextString(m) }
func (*Monitoring_MonitoringDestination) ProtoMessage()    {}
func (*Monitoring_MonitoringDestination) Descriptor() ([]byte, []int) {
	return fileDescriptor14, []int{0, 0}
}

func init() {
	proto.RegisterType((*Monitoring)(nil), "google.api.Monitoring")
	proto.RegisterType((*Monitoring_MonitoringDestination)(nil), "google.api.Monitoring.MonitoringDestination")
}

func init() { proto.RegisterFile("google/api/monitoring.proto", fileDescriptor14) }

var fileDescriptor14 = []byte{
	// 226 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x9c, 0x90, 0x4f, 0x4a, 0xc4, 0x30,
	0x14, 0xc6, 0xc9, 0x08, 0xca, 0x3c, 0x41, 0x31, 0x38, 0x50, 0x46, 0x17, 0x83, 0x6e, 0x66, 0xa1,
	0x29, 0xe8, 0x0d, 0x8a, 0x5b, 0xa1, 0xe4, 0x02, 0x35, 0xa6, 0xa1, 0x04, 0x4c, 0x5e, 0x48, 0xd2,
	0x9b, 0x79, 0x40, 0xa9, 0x69, 0x9b, 0x20, 0xae, 0x66, 0xd7, 0xe6, 0xfb, 0xf3, 0x7b, 0x7c, 0x70,
	0x37, 0x20, 0x0e, 0x5f, 0xaa, 0x16, 0x4e, 0xd7, 0x06, 0xad, 0x8e, 0xe8, 0xb5, 0x1d, 0x98, 0xf3,
	0x18, 0x91, 0x42, 0x12, 0x99, 0x70, 0x7a, 0x7f, 0x5f, 0x18, 0x85, 0xb5, 0x18, 0x45, 0xd4, 0x68,
	0x43, 0x72, 0x3e, 0x7c, 0x6f, 0x00, 0xde, 0xd7, 0x38, 0x15, 0xb0, 0x73, 0x1e, 0xfb, 0x51, 0x2a,
	0xdf, 0xf5, 0x2a, 0x44, 0x6d, 0x93, 0xbb, 0x22, 0x87, 0xb3, 0xe3, 0xe5, 0xcb, 0x13, 0xcb, 0xc5,
	0x2c, 0xc7, 0x8a, 0xcf, 0xb7, 0x1c, 0xe2, 0xb7, 0x4b, 0x55, 0xf1, 0x18, 0x26, 0x84, 0x44, 0x1b,
	0x46, 0xf3, 0x17, 0xb1, 0x39, 0x05, 0xb1, 0x54, 0x95, 0x88, 0xfd, 0x07, 0xec, 0xfe, 0xb5, 0xd3,
	0x67, 0xa0, 0xf3, 0x56, 0xaa, 0xef, 0xbc, 0x0a, 0x38, 0x7a, 0xa9, 0x2a, 0x72, 0x20, 0xc7, 0x2d,
	0xbf, 0x59, 0x15, 0x3e, 0x0b, 0xb4, 0x82, 0x0b, 0xa3, 0xa2, 0xd7, 0x32, 0x1d, 0xb7, 0xe5, 0xcb,
	0x6f, 0xf3, 0x08, 0x57, 0x12, 0x4d, 0x71, 0x6a, 0x73, 0x9d, 0x89, 0xed, 0xb4, 0x6c, 0x4b, 0x3e,
	0xcf, 0x7f, 0x27, 0x7e, 0xfd, 0x09, 0x00, 0x00, 0xff, 0xff, 0xb6, 0xd0, 0x8d, 0x3b, 0xab, 0x01,
	0x00, 0x00,
}
