package glog

import (
	"encoding/json"
	"github.com/golang/glog"
	structpb "github.com/golang/protobuf/ptypes/struct"
	sc "google/api/servicecontrol/v1"
	"somnacin-internal/mixologist/mixologist"
	"time"
)

const (
	Name = "glog"
)

func init() {
	mixologist.RegisterReportConsumer(Name, new(builder))
}

type (
	consumer struct{}
	builder  struct{}

	Resource struct {
		// type. should always be API.
		Type   string            `json:"type,omitempty"`
		Labels map[string]string `json:"labels,omitempty"`
	}

	LogEntry struct {
		Name         string                 `json:"logName,omitempty"`
		Timestamp    time.Time              `json:"timestamp,omitempty"`
		Resource     Resource               `json:"resource,omitempty"`
		Labels       map[string]string      `json:"labels,omitempty"`
		Severity     string                 `json:"severity,omitempty"`
		ProtoPayload string                 `json:"protoPayload,omitempty"`
		TextPayload  string                 `json:"textPayload,omitempty"`
		JSONPayload  map[string]interface{} `json:"jsonPayload,omitempty"`
	}
)

func convert(k *structpb.Value) interface{} {
	kind := k.GetKind()
	switch kind.(type) {
	case *structpb.Value_StringValue:
		return k.GetStringValue()
	case *structpb.Value_NumberValue:
		return k.GetNumberValue()
	case *structpb.Value_BoolValue:
		return k.GetBoolValue()
	case *structpb.Value_ListValue:
		t := []interface{}{}
		for _, v := range k.GetListValue().GetValues() {
			t = append(t, convert(v))
		}
		return t
	case *structpb.Value_StructValue:
		m := map[string]interface{}{}
		for k, v := range k.GetStructValue().GetFields() {
			m[k] = convert(v)
		}
		return m
	default:
		// TODO(dougreid): is this what we want here?
		return nil
	}

}

func jsonPayload(s *structpb.Struct) map[string]interface{} {
	tmp := map[string]interface{}{}
	for k, v := range s.Fields {
		tmp[k] = convert(v)
	}
	return tmp
}

func apiResource(labels map[string]string) Resource {
	rl := map[string]string{}
	for _, v := range mixologist.MonitoredAPIResourceLabels {
		if val, ok := labels[v]; ok {
			rl[v] = val
		}
	}
	return Resource{
		Type:   "api",
		Labels: rl,
	}
}

// Consume -- Called to consume 1 reportMsg at a time
func (c *consumer) Consume(reportMsg *sc.ReportRequest) (err error) {
	svc := reportMsg.ServiceName
	for _, oprn := range reportMsg.GetOperations() {
		defaultLabels := oprn.GetLabels()
		defaultLabels[mixologist.CloudService] = svc
		defaultLabels[mixologist.ConsumerID] = oprn.ConsumerId

		for _, le := range oprn.GetLogEntries() {
			ts := oprn.StartTime
			if le.Timestamp != nil {
				ts = le.Timestamp
			}
			t := time.Unix(ts.Seconds, int64(ts.Nanos))

			entry := LogEntry{
				Name:      le.Name,
				Resource:  apiResource(defaultLabels),
				Severity:  le.Severity.String(),
				Timestamp: t,
				Labels:    map[string]string{}, // TODO(dougreid): populate
			}

			if pp := le.GetProtoPayload(); pp != nil {
				// TODO(dougreid): need some plan to handle proto payload here? maybe post-prototype...
				// leave untested with basic String() call for now
				entry.ProtoPayload = pp.String()
			}

			if tp := le.GetTextPayload(); tp != "" {
				entry.TextPayload = tp
			}

			if sp := le.GetStructPayload(); sp != nil {
				entry.JSONPayload = jsonPayload(sp)
			}

			var out []byte
			if out, err = json.Marshal(entry); err != nil {
				// TODO(dougreid): do we want to err out here?
				return nil
			}
			glog.Infof("%s", out)
		}
	}
	glog.Flush()

	//TODO return error when it makes sense
	return nil
}

// GetName interface method
func (c *consumer) GetName() string {
	return Name
}

// Not needed.
func (c *consumer) GetPrefixAndHandler() *mixologist.PrefixAndHandler {
	return nil
}

func (b *builder) NewConsumer(c mixologist.Config) mixologist.ReportConsumer {
	return &consumer{}
}
