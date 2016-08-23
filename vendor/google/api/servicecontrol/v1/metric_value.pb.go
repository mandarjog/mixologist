// Code generated by protoc-gen-go.
// source: google/api/servicecontrol/v1/metric_value.proto
// DO NOT EDIT!

package servicecontrol

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf3 "github.com/golang/protobuf/ptypes/timestamp"
import google_type "google/type"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Represents a single metric value.
type MetricValue struct {
	// The labels describing the metric value.
	// See comments on Operation.labels for the overriding relationship.
	Labels map[string]string `protobuf:"bytes,1,rep,name=labels" json:"labels,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	// The start of the time period over which this metric value's measurement
	// applies. The time period has different semantics for different metric
	// types (cumulative, delta, and gauge). See the metric definition
	// documentation in the service configuration for details.
	StartTime *google_protobuf3.Timestamp `protobuf:"bytes,2,opt,name=start_time,json=startTime" json:"start_time,omitempty"`
	// The end of the time period over which this metric value's measurement
	// applies.
	EndTime *google_protobuf3.Timestamp `protobuf:"bytes,3,opt,name=end_time,json=endTime" json:"end_time,omitempty"`
	// The value. The type of value used in the request must
	// agree with the metric definition in the service configuration, otherwise
	// the MetricValue is rejected.
	//
	// Types that are valid to be assigned to Value:
	//	*MetricValue_BoolValue
	//	*MetricValue_Int64Value
	//	*MetricValue_DoubleValue
	//	*MetricValue_StringValue
	//	*MetricValue_DistributionValue
	//	*MetricValue_MoneyValue
	Value isMetricValue_Value `protobuf_oneof:"value"`
}

func (m *MetricValue) Reset()                    { *m = MetricValue{} }
func (m *MetricValue) String() string            { return proto.CompactTextString(m) }
func (*MetricValue) ProtoMessage()               {}
func (*MetricValue) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{0} }

type isMetricValue_Value interface {
	isMetricValue_Value()
}

type MetricValue_BoolValue struct {
	BoolValue bool `protobuf:"varint,4,opt,name=bool_value,json=boolValue,oneof"`
}
type MetricValue_Int64Value struct {
	Int64Value int64 `protobuf:"varint,5,opt,name=int64_value,json=int64Value,oneof"`
}
type MetricValue_DoubleValue struct {
	DoubleValue float64 `protobuf:"fixed64,6,opt,name=double_value,json=doubleValue,oneof"`
}
type MetricValue_StringValue struct {
	StringValue string `protobuf:"bytes,7,opt,name=string_value,json=stringValue,oneof"`
}
type MetricValue_DistributionValue struct {
	DistributionValue *Distribution `protobuf:"bytes,8,opt,name=distribution_value,json=distributionValue,oneof"`
}
type MetricValue_MoneyValue struct {
	MoneyValue *google_type.Money `protobuf:"bytes,9,opt,name=money_value,json=moneyValue,oneof"`
}

func (*MetricValue_BoolValue) isMetricValue_Value()         {}
func (*MetricValue_Int64Value) isMetricValue_Value()        {}
func (*MetricValue_DoubleValue) isMetricValue_Value()       {}
func (*MetricValue_StringValue) isMetricValue_Value()       {}
func (*MetricValue_DistributionValue) isMetricValue_Value() {}
func (*MetricValue_MoneyValue) isMetricValue_Value()        {}

func (m *MetricValue) GetValue() isMetricValue_Value {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *MetricValue) GetLabels() map[string]string {
	if m != nil {
		return m.Labels
	}
	return nil
}

func (m *MetricValue) GetStartTime() *google_protobuf3.Timestamp {
	if m != nil {
		return m.StartTime
	}
	return nil
}

func (m *MetricValue) GetEndTime() *google_protobuf3.Timestamp {
	if m != nil {
		return m.EndTime
	}
	return nil
}

func (m *MetricValue) GetBoolValue() bool {
	if x, ok := m.GetValue().(*MetricValue_BoolValue); ok {
		return x.BoolValue
	}
	return false
}

func (m *MetricValue) GetInt64Value() int64 {
	if x, ok := m.GetValue().(*MetricValue_Int64Value); ok {
		return x.Int64Value
	}
	return 0
}

func (m *MetricValue) GetDoubleValue() float64 {
	if x, ok := m.GetValue().(*MetricValue_DoubleValue); ok {
		return x.DoubleValue
	}
	return 0
}

func (m *MetricValue) GetStringValue() string {
	if x, ok := m.GetValue().(*MetricValue_StringValue); ok {
		return x.StringValue
	}
	return ""
}

func (m *MetricValue) GetDistributionValue() *Distribution {
	if x, ok := m.GetValue().(*MetricValue_DistributionValue); ok {
		return x.DistributionValue
	}
	return nil
}

func (m *MetricValue) GetMoneyValue() *google_type.Money {
	if x, ok := m.GetValue().(*MetricValue_MoneyValue); ok {
		return x.MoneyValue
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*MetricValue) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _MetricValue_OneofMarshaler, _MetricValue_OneofUnmarshaler, _MetricValue_OneofSizer, []interface{}{
		(*MetricValue_BoolValue)(nil),
		(*MetricValue_Int64Value)(nil),
		(*MetricValue_DoubleValue)(nil),
		(*MetricValue_StringValue)(nil),
		(*MetricValue_DistributionValue)(nil),
		(*MetricValue_MoneyValue)(nil),
	}
}

func _MetricValue_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*MetricValue)
	// value
	switch x := m.Value.(type) {
	case *MetricValue_BoolValue:
		t := uint64(0)
		if x.BoolValue {
			t = 1
		}
		b.EncodeVarint(4<<3 | proto.WireVarint)
		b.EncodeVarint(t)
	case *MetricValue_Int64Value:
		b.EncodeVarint(5<<3 | proto.WireVarint)
		b.EncodeVarint(uint64(x.Int64Value))
	case *MetricValue_DoubleValue:
		b.EncodeVarint(6<<3 | proto.WireFixed64)
		b.EncodeFixed64(math.Float64bits(x.DoubleValue))
	case *MetricValue_StringValue:
		b.EncodeVarint(7<<3 | proto.WireBytes)
		b.EncodeStringBytes(x.StringValue)
	case *MetricValue_DistributionValue:
		b.EncodeVarint(8<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.DistributionValue); err != nil {
			return err
		}
	case *MetricValue_MoneyValue:
		b.EncodeVarint(9<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.MoneyValue); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("MetricValue.Value has unexpected type %T", x)
	}
	return nil
}

func _MetricValue_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*MetricValue)
	switch tag {
	case 4: // value.bool_value
		if wire != proto.WireVarint {
			return true, proto.ErrInternalBadWireType
		}
		x, err := b.DecodeVarint()
		m.Value = &MetricValue_BoolValue{x != 0}
		return true, err
	case 5: // value.int64_value
		if wire != proto.WireVarint {
			return true, proto.ErrInternalBadWireType
		}
		x, err := b.DecodeVarint()
		m.Value = &MetricValue_Int64Value{int64(x)}
		return true, err
	case 6: // value.double_value
		if wire != proto.WireFixed64 {
			return true, proto.ErrInternalBadWireType
		}
		x, err := b.DecodeFixed64()
		m.Value = &MetricValue_DoubleValue{math.Float64frombits(x)}
		return true, err
	case 7: // value.string_value
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		x, err := b.DecodeStringBytes()
		m.Value = &MetricValue_StringValue{x}
		return true, err
	case 8: // value.distribution_value
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(Distribution)
		err := b.DecodeMessage(msg)
		m.Value = &MetricValue_DistributionValue{msg}
		return true, err
	case 9: // value.money_value
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(google_type.Money)
		err := b.DecodeMessage(msg)
		m.Value = &MetricValue_MoneyValue{msg}
		return true, err
	default:
		return false, nil
	}
}

func _MetricValue_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*MetricValue)
	// value
	switch x := m.Value.(type) {
	case *MetricValue_BoolValue:
		n += proto.SizeVarint(4<<3 | proto.WireVarint)
		n += 1
	case *MetricValue_Int64Value:
		n += proto.SizeVarint(5<<3 | proto.WireVarint)
		n += proto.SizeVarint(uint64(x.Int64Value))
	case *MetricValue_DoubleValue:
		n += proto.SizeVarint(6<<3 | proto.WireFixed64)
		n += 8
	case *MetricValue_StringValue:
		n += proto.SizeVarint(7<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(len(x.StringValue)))
		n += len(x.StringValue)
	case *MetricValue_DistributionValue:
		s := proto.Size(x.DistributionValue)
		n += proto.SizeVarint(8<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *MetricValue_MoneyValue:
		s := proto.Size(x.MoneyValue)
		n += proto.SizeVarint(9<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

// Represents a set of metric values in the same metric.
// Each metric value in the set should have a unique combination of start time,
// end time, and label values.
type MetricValueSet struct {
	// The metric name defined in the service configuration.
	MetricName string `protobuf:"bytes,1,opt,name=metric_name,json=metricName" json:"metric_name,omitempty"`
	// The values in this metric.
	MetricValues []*MetricValue `protobuf:"bytes,2,rep,name=metric_values,json=metricValues" json:"metric_values,omitempty"`
}

func (m *MetricValueSet) Reset()                    { *m = MetricValueSet{} }
func (m *MetricValueSet) String() string            { return proto.CompactTextString(m) }
func (*MetricValueSet) ProtoMessage()               {}
func (*MetricValueSet) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{1} }

func (m *MetricValueSet) GetMetricValues() []*MetricValue {
	if m != nil {
		return m.MetricValues
	}
	return nil
}

func init() {
	proto.RegisterType((*MetricValue)(nil), "google.api.servicecontrol.v1.MetricValue")
	proto.RegisterType((*MetricValueSet)(nil), "google.api.servicecontrol.v1.MetricValueSet")
}

func init() { proto.RegisterFile("google/api/servicecontrol/v1/metric_value.proto", fileDescriptor3) }

var fileDescriptor3 = []byte{
	// 465 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x8c, 0x93, 0x4f, 0x8f, 0xd3, 0x30,
	0x10, 0xc5, 0xeb, 0x76, 0xb7, 0x7f, 0x26, 0x0b, 0x02, 0x83, 0x44, 0x54, 0x21, 0x35, 0x2c, 0x97,
	0xc0, 0xc1, 0xd1, 0x2e, 0x14, 0xb1, 0x88, 0x53, 0x05, 0x52, 0x0f, 0x74, 0xb5, 0x0a, 0x88, 0x0b,
	0x87, 0x95, 0xd3, 0x0e, 0x95, 0x45, 0x6c, 0x47, 0x89, 0x5b, 0xa9, 0x47, 0xbe, 0x13, 0x1f, 0x8e,
	0x23, 0xf2, 0x9f, 0x42, 0x7a, 0x29, 0x7b, 0xcb, 0x8c, 0x7f, 0xef, 0xe9, 0xd9, 0x33, 0x81, 0x6c,
	0xad, 0xf5, 0xba, 0xc4, 0x8c, 0x57, 0x22, 0x6b, 0xb0, 0xde, 0x8a, 0x25, 0x2e, 0xb5, 0x32, 0xb5,
	0x2e, 0xb3, 0xed, 0x45, 0x26, 0xd1, 0xd4, 0x62, 0x79, 0xbb, 0xe5, 0xe5, 0x06, 0x59, 0x55, 0x6b,
	0xa3, 0xe9, 0x53, 0x2f, 0x60, 0xbc, 0x12, 0xec, 0x50, 0xc0, 0xb6, 0x17, 0xe3, 0xe3, 0x76, 0x2b,
	0xd1, 0x98, 0x5a, 0x14, 0x1b, 0x23, 0xb4, 0xf2, 0x76, 0xe3, 0x49, 0x10, 0xb8, 0xaa, 0xd8, 0x7c,
	0xcf, 0x8c, 0x90, 0xd8, 0x18, 0x2e, 0xab, 0x00, 0x3c, 0x09, 0x80, 0xd9, 0x55, 0x98, 0x49, 0xad,
	0x70, 0xe7, 0x0f, 0xce, 0x7f, 0x9d, 0x40, 0xb4, 0x70, 0xf9, 0xbe, 0xda, 0x78, 0x74, 0x01, 0xfd,
	0x92, 0x17, 0x58, 0x36, 0x31, 0x49, 0x7a, 0x69, 0x74, 0x39, 0x65, 0xc7, 0x92, 0xb2, 0x96, 0x94,
	0x7d, 0x72, 0xba, 0x8f, 0xca, 0xd4, 0xbb, 0x3c, 0x98, 0xd0, 0x2b, 0x80, 0xc6, 0xf0, 0xda, 0xdc,
	0xda, 0x40, 0x71, 0x37, 0x21, 0x69, 0x74, 0x39, 0xde, 0x5b, 0xee, 0xd3, 0xb2, 0x2f, 0xfb, 0xb4,
	0xf9, 0xc8, 0xd1, 0xb6, 0xa6, 0x53, 0x18, 0xa2, 0x5a, 0x79, 0x61, 0xef, 0xbf, 0xc2, 0x01, 0xaa,
	0x95, 0x93, 0x4d, 0x00, 0x0a, 0xad, 0x4b, 0xff, 0xda, 0xf1, 0x49, 0x42, 0xd2, 0xe1, 0xbc, 0x93,
	0x8f, 0x6c, 0xcf, 0xdf, 0xf0, 0x19, 0x44, 0x42, 0x99, 0x37, 0xaf, 0x03, 0x71, 0x9a, 0x90, 0xb4,
	0x37, 0xef, 0xe4, 0xe0, 0x9a, 0x1e, 0x79, 0x0e, 0x67, 0x2b, 0xbd, 0x29, 0x4a, 0x0c, 0x4c, 0x3f,
	0x21, 0x29, 0x99, 0x77, 0xf2, 0xc8, 0x77, 0xff, 0x42, 0x76, 0x0e, 0x6a, 0x1d, 0xa0, 0x41, 0x42,
	0xd2, 0x91, 0x85, 0x7c, 0xd7, 0x43, 0xdf, 0x80, 0xb6, 0xc7, 0x15, 0xd0, 0xa1, 0xbb, 0xce, 0xcb,
	0xe3, 0x4f, 0xfb, 0xa1, 0xa5, 0x9b, 0x77, 0xf2, 0x87, 0x6d, 0x1f, 0x6f, 0x3e, 0x85, 0xc8, 0x8d,
	0x32, 0xb8, 0x8e, 0x9c, 0x2b, 0xdd, 0xbb, 0xda, 0x51, 0xb3, 0x85, 0x3d, 0xb7, 0xb7, 0x73, 0xa0,
	0x93, 0x8d, 0xaf, 0x20, 0x6a, 0x8d, 0x8a, 0x3e, 0x80, 0xde, 0x0f, 0xdc, 0xc5, 0xc4, 0xc6, 0xcf,
	0xed, 0x27, 0x7d, 0x0c, 0xa7, 0xde, 0xb1, 0xeb, 0x7a, 0xbe, 0x78, 0xd7, 0x7d, 0x4b, 0x66, 0x83,
	0x70, 0x72, 0xfe, 0x93, 0xc0, 0xfd, 0xd6, 0xec, 0x3f, 0xa3, 0xa1, 0x13, 0x88, 0xc2, 0xa2, 0x2b,
	0x2e, 0x31, 0xf8, 0x81, 0x6f, 0x5d, 0x73, 0x89, 0xf4, 0x1a, 0xee, 0xb5, 0xff, 0x84, 0x26, 0xee,
	0xba, 0x0d, 0x7b, 0x71, 0xe7, 0x0d, 0xcb, 0xcf, 0xe4, 0xbf, 0xa2, 0x99, 0xbd, 0x87, 0x64, 0xa9,
	0xe5, 0x51, 0xf5, 0xec, 0xd1, 0x61, 0xc8, 0x1b, 0xbb, 0x38, 0x37, 0xe4, 0x37, 0x21, 0x45, 0xdf,
	0x2d, 0xd1, 0xab, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x84, 0x59, 0xd1, 0xcd, 0xbb, 0x03, 0x00,
	0x00,
}
