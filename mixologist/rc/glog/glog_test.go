package glog

import (
	"bytes"
	"flag"
	"github.com/golang/glog"
	structpb "github.com/golang/protobuf/ptypes/struct"
	tspb "github.com/golang/protobuf/ptypes/timestamp"
	sc "google/api/servicecontrol/v1"
	"google/logging/type"
	"io"
	"os"
	"somnacin-internal/mixologist/mixologist"
	"strings"
	"testing"
)

var (
	testLabels = map[string]string{
		mixologist.CloudService:  "test-service",
		mixologist.CloudLocation: "us-west1-a",
		mixologist.ConsumerID:    "project:test-project",
		mixologist.APIMethod:     "TestMethod",
		mixologist.APIVersion:    "test-service-v1.appspot.com",
	}
	testSvc      = "test-service"
	testTime     = &tspb.Timestamp{Seconds: 1471970653, Nanos: 808341000}
	testLog      = "endpoints_log"
	testSeverity = _type.LogSeverity_INFO
)

func newReportReq(les ...*sc.LogEntry) *sc.ReportRequest {
	return &sc.ReportRequest{ServiceName: testSvc, Operations: []*sc.Operation{newOperation(les...)}}
}

func newOperation(les ...*sc.LogEntry) *sc.Operation {
	return &sc.Operation{Labels: testLabels, LogEntries: les}
}

func newStructLogEntry(payload *sc.LogEntry_StructPayload) *sc.LogEntry {
	return &sc.LogEntry{
		Name:      testLog,
		Severity:  testSeverity,
		Timestamp: testTime,
		Payload:   payload,
	}
}

func newTextLogEntry(s string) *sc.LogEntry {
	return &sc.LogEntry{
		Name:      testLog,
		Severity:  testSeverity,
		Timestamp: testTime,
		Payload:   &sc.LogEntry_TextPayload{TextPayload: s},
	}
}

func newStructPayload(m map[string]*structpb.Value) *sc.LogEntry_StructPayload {
	return &sc.LogEntry_StructPayload{StructPayload: &structpb.Struct{Fields: m}}
}

func stringVal(val string) *structpb.Value {
	return &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: val}}
}

func boolVal(b bool) *structpb.Value {
	return &structpb.Value{Kind: &structpb.Value_BoolValue{BoolValue: b}}
}

func floatListVal(item float64) *structpb.Value {
	return &structpb.Value{Kind: &structpb.Value_ListValue{ListValue: &structpb.ListValue{Values: []*structpb.Value{&structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: item}}}}}}
}

func structVal(m map[string]*structpb.Value) *structpb.Value {
	return &structpb.Value{Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{Fields: m}}}
}

func TestConsume(t *testing.T) {
	var consumeTests = []struct {
		name      string
		report    *sc.ReportRequest
		wantLines int64
		wantText  []string
	}{
		{
			name:      "Text Payload, Single Entry",
			report:    newReportReq(newTextLogEntry("some text")),
			wantLines: 1,
			wantText:  []string{"{\"logName\":\"endpoints_log\",\"timestamp\":\"2016-08-23T09:44:13.808341-07:00\",\"resource\":{\"type\":\"api\",\"labels\":{\"cloud.googleapis.com/location\":\"us-west1-a\",\"cloud.googleapis.com/service\":\"test-service\",\"serviceruntime.googleapis.com/api_method\":\"TestMethod\",\"serviceruntime.googleapis.com/api_version\":\"test-service-v1.appspot.com\"}},\"severity\":\"INFO\",\"textPayload\":\"some text\"}"},
		},
		{
			name:      "Struct Payload (String), Single Entry",
			report:    newReportReq(newStructLogEntry(newStructPayload(map[string]*structpb.Value{"api_method": stringVal("ListShelves")}))),
			wantLines: 1,
			wantText:  []string{"{\"logName\":\"endpoints_log\",\"timestamp\":\"2016-08-23T09:44:13.808341-07:00\",\"resource\":{\"type\":\"api\",\"labels\":{\"cloud.googleapis.com/location\":\"us-west1-a\",\"cloud.googleapis.com/service\":\"test-service\",\"serviceruntime.googleapis.com/api_method\":\"TestMethod\",\"serviceruntime.googleapis.com/api_version\":\"test-service-v1.appspot.com\"}},\"severity\":\"INFO\",\"jsonPayload\":{\"api_method\":\"ListShelves\"}}"},
		},
		{
			name:      "Struct Payload (String), Single Entry (with bool)",
			report:    newReportReq(newStructLogEntry(newStructPayload(map[string]*structpb.Value{"boolean": boolVal(true)}))),
			wantLines: 1,
			wantText:  []string{"{\"logName\":\"endpoints_log\",\"timestamp\":\"2016-08-23T09:44:13.808341-07:00\",\"resource\":{\"type\":\"api\",\"labels\":{\"cloud.googleapis.com/location\":\"us-west1-a\",\"cloud.googleapis.com/service\":\"test-service\",\"serviceruntime.googleapis.com/api_method\":\"TestMethod\",\"serviceruntime.googleapis.com/api_version\":\"test-service-v1.appspot.com\"}},\"severity\":\"INFO\",\"jsonPayload\":{\"boolean\":true}}"},
		},
		{
			name:      "Struct Payload (String), Single Entry (with list)",
			report:    newReportReq(newStructLogEntry(newStructPayload(map[string]*structpb.Value{"latency": floatListVal(78.234)}))),
			wantLines: 1,
			wantText:  []string{"{\"logName\":\"endpoints_log\",\"timestamp\":\"2016-08-23T09:44:13.808341-07:00\",\"resource\":{\"type\":\"api\",\"labels\":{\"cloud.googleapis.com/location\":\"us-west1-a\",\"cloud.googleapis.com/service\":\"test-service\",\"serviceruntime.googleapis.com/api_method\":\"TestMethod\",\"serviceruntime.googleapis.com/api_version\":\"test-service-v1.appspot.com\"}},\"severity\":\"INFO\",\"jsonPayload\":{\"latency\":[78.234]}}"},
		},
		{
			name:      "Struct Payload (Struct-ception), Single Entry",
			report:    newReportReq(newStructLogEntry(newStructPayload(map[string]*structpb.Value{"embedded": structVal(map[string]*structpb.Value{"test": stringVal("test")})}))),
			wantLines: 1,
			wantText: []string{
				"{\"logName\":\"endpoints_log\",\"timestamp\":\"2016-08-23T09:44:13.808341-07:00\",\"resource\":{\"type\":\"api\",\"labels\":{\"cloud.googleapis.com/location\":\"us-west1-a\",\"cloud.googleapis.com/service\":\"test-service\",\"serviceruntime.googleapis.com/api_method\":\"TestMethod\",\"serviceruntime.googleapis.com/api_version\":\"test-service-v1.appspot.com\"}},\"severity\":\"INFO\",\"jsonPayload\":{\"embedded\":{\"test\":\"test\"}}}",
			},
		},
		{
			name: "Struct Payload (String), Multiple Entries",
			report: newReportReq(newStructLogEntry(newStructPayload(map[string]*structpb.Value{"api_method": stringVal("ListShelves")})),
				newStructLogEntry(newStructPayload(map[string]*structpb.Value{"url": stringVal("/shelves")}))),
			wantLines: 2,
			wantText: []string{
				"{\"logName\":\"endpoints_log\",\"timestamp\":\"2016-08-23T09:44:13.808341-07:00\",\"resource\":{\"type\":\"api\",\"labels\":{\"cloud.googleapis.com/location\":\"us-west1-a\",\"cloud.googleapis.com/service\":\"test-service\",\"serviceruntime.googleapis.com/api_method\":\"TestMethod\",\"serviceruntime.googleapis.com/api_version\":\"test-service-v1.appspot.com\"}},\"severity\":\"INFO\",\"jsonPayload\":{\"api_method\":\"ListShelves\"}}",
				"{\"logName\":\"endpoints_log\",\"timestamp\":\"2016-08-23T09:44:13.808341-07:00\",\"resource\":{\"type\":\"api\",\"labels\":{\"cloud.googleapis.com/location\":\"us-west1-a\",\"cloud.googleapis.com/service\":\"test-service\",\"serviceruntime.googleapis.com/api_method\":\"TestMethod\",\"serviceruntime.googleapis.com/api_version\":\"test-service-v1.appspot.com\"}},\"severity\":\"INFO\",\"jsonPayload\":{\"url\":\"/shelves\"}}",
			},
		},
	}

	b := &builder{}
	c := b.NewConsumer(mixologist.Config{})

	// we want glog to go to stderr so we can redirect for
	// test validation of log generation
	flag.Set("logtostderr", "true")

	for _, v := range consumeTests {

		old := os.Stderr // for restore
		r, w, _ := os.Pipe()
		os.Stderr = w // redirecting

		// copy over the output from stderr
		outC := make(chan string)
		go func() {
			var buf bytes.Buffer
			io.Copy(&buf, r)
			outC <- buf.String()
		}()

		sl := glog.Stats.Info.Lines() // validate lines

		c.Consume(v.report)

		if got := glog.Stats.Info.Lines() - sl; got != v.wantLines {
			t.Errorf("%s: got %v lines, want %v", v.name, v.wantLines)
		}

		// back to normal state
		w.Close()
		os.Stderr = old
		out := <-outC

		for i, s := range strings.Split(out, "\n") {
			if s == "" {
				continue
			}
			if got := trim(s); got != v.wantText[i] {
				t.Errorf("%s: got %v, want %v", v.name, got, v.wantText[i])
			}
		}
	}
}

// trims glog line prefix stuff off of output log lines (glog prefix ends with ']')
func trim(s string) string {
	return strings.TrimRight(strings.TrimLeft(strings.SplitN(s, "]", 2)[1], " "), "\n")
}