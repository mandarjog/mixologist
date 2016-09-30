package logsAdapter

import (
	"github.com/golang/glog"
	structpb "github.com/golang/protobuf/ptypes/struct"
	sc "google/api/servicecontrol/v1"
	"io"
	"os"
	"somnacin-internal/mixologist/mixologist"
	"time"
)

const (
	Name = "mixologist.io/consumers/logsAdapter"
)

var (
	plugins = make(map[string]mixologist.LoggerBuilder)
)

func init() {
	mixologist.RegisterReportConsumer(Name, new(builder))
}

// RegisterLogsSink should only be called in init() methods.
func RegisterLogsSink(name string, b mixologist.LoggerBuilder) {
	// TODO(dougreid): worry about collisions?
	plugins[name] = b
}

type (
	consumer struct {
		loggers []mixologist.Logger
	}
	builder struct{}
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

func structPayload(s *structpb.Struct) map[string]interface{} {
	tmp := map[string]interface{}{}
	for k, v := range s.Fields {
		tmp[k] = convert(v)
	}
	return tmp
}

func apiResource(labels map[string]string) mixologist.Resource {
	rl := map[string]string{}
	for _, v := range mixologist.MonitoredAPIResourceLabels {
		if val, ok := labels[v]; ok {
			rl[v] = val
		}
	}
	return mixologist.Resource{
		Type:   "api",
		Labels: rl,
	}
}

func logEntry(src *sc.LogEntry, lbls map[string]string, t time.Time) mixologist.LogEntry {
	entry := mixologist.LogEntry{
		Name:      src.Name,
		Resource:  apiResource(lbls),
		ID:        src.InsertId,
		Severity:  src.Severity.String(),
		Timestamp: t,
		Labels:    map[string]string{}, // TODO(dougreid): populate
	}

	if pp := src.GetProtoPayload(); pp != nil {
		// TODO(dougreid): need some plan to handle proto payload here? maybe post-prototype...
		// leave untested with basic String() call for now
		entry.ProtoPayload = pp.String()
	}

	if tp := src.GetTextPayload(); tp != "" {
		entry.TextPayload = tp
	}

	if sp := src.GetStructPayload(); sp != nil {
		entry.StructPayload = structPayload(sp)
	}

	return entry
}

func startTime(o *sc.Operation, l *sc.LogEntry) time.Time {
	s := o.StartTime
	if l.Timestamp != nil {
		s = l.Timestamp
	}
	return time.Unix(s.Seconds, int64(s.Nanos))
}

// Consume -- Called to consume 1 reportMsg at a time
func (c *consumer) Consume(reportMsg *sc.ReportRequest) (err error) {
	svc := reportMsg.ServiceName
	for _, oprn := range reportMsg.GetOperations() {
		defaultLabels := oprn.GetLabels()
		oid := oprn.OperationId
		defaultLabels[mixologist.CloudService] = svc
		defaultLabels[mixologist.ConsumerID] = oprn.ConsumerId

		glog.Infof("logs adapter default labels: %v", defaultLabels)

		for _, le := range oprn.GetLogEntries() {
			entry := logEntry(le, defaultLabels, startTime(oprn, le))
			entry.OperationID = oid
			for _, v := range c.loggers {
				if err := v.Log(entry); err != nil {
					glog.Errorf("could not log to logger %s: %v", v.Name, err)
				}
			}
		}
	}

	for _, v := range c.loggers {
		v.Flush()
	}

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

func (b *builder) BuildConsumer(c mixologist.Config) (mixologist.ReportConsumer, error) {
	glog.Infof("adding consumer with config: %v", c)
	cons := &consumer{}
	for _, be := range c.Logging.Backends {
		if v, ok := plugins[be]; ok {
			cons.loggers = append(cons.loggers, v.Build(c))
		}
	}
	if c.Logging.UseDefault {
		cons.loggers = append(cons.loggers, stdLogger{})
	}
	return cons, nil
}

// default logger builder
type stdLoggerBuilder struct{}

func (b stdLoggerBuilder) Build(c mixologist.Config) mixologist.Logger { return stdLogger{} }

// default logger

type stdLogger struct{}

func (l stdLogger) Name() string { return "mixologist.io/loggers/default" }
func (l stdLogger) Log(le mixologist.LogEntry) error {
	out, err := mixologist.JSONBytes(le)
	if err != nil {
		return err
	}

	// for now, default to always writing to stderr
	_, err = l.write(os.Stderr, out)
	return err
}
func (l stdLogger) Flush() {}

func (l stdLogger) write(w io.Writer, b []byte) (int, error) { return w.Write(b) }
