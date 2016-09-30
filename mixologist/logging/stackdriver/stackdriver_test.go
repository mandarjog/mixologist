package stackdriver

import (
	"errors"
	"github.com/golang/protobuf/proto"
	structpb "github.com/golang/protobuf/ptypes/struct"
	timestamppb "github.com/golang/protobuf/ptypes/timestamp"
	"golang.org/x/net/context"
	resourcepb "google.golang.org/genproto/googleapis/api/monitoredres"
	loggingtypepb "google.golang.org/genproto/googleapis/logging/type"
	loggingpb "google.golang.org/genproto/googleapis/logging/v2"
	"somnacin-internal/mixologist/mixologist"
	"testing"
	"time"
)

type fakeClient struct {
	returnErr bool
	lastReq   *loggingpb.WriteLogEntriesRequest
}

func (f *fakeClient) WriteLogEntries(ctx context.Context, req *loggingpb.WriteLogEntriesRequest) (*loggingpb.WriteLogEntriesResponse, error) {
	if f.returnErr {
		f.lastReq = nil
		return nil, errors.New("error writing logs to cloud logging")
	}
	f.lastReq = req
	return &loggingpb.WriteLogEntriesResponse{}, nil
}

func TestLog(t *testing.T) {
	tests := []struct {
		name      string
		in        mixologist.LogEntry
		want      *loggingpb.LogEntry
		reportErr bool
	}{
		{
			name: "Log Entry with List",
			in: mixologist.LogEntry{
				Name:          "test-log",
				Resource:      mixologist.Resource{Type: "api", Labels: map[string]string{mixologist.APIMethod: "Test", mixologist.CloudLocation: "us-west-1b", mixologist.CloudProject: "test-project", mixologist.CloudService: "test-service", mixologist.APIVersion: "v2"}},
				Labels:        map[string]string{"http_method": "GET"},
				Severity:      "INFO",
				OperationID:   "operation-id",
				Timestamp:     time.Date(2016, 9, 1, 11, 15, 59, 0, time.UTC),
				StructPayload: map[string]interface{}{"producer_project_id": "test-project"},
			},
			want: &loggingpb.LogEntry{
				LogName: "projects/test-project/logs/test-log",
				Resource: &resourcepb.MonitoredResource{
					Type:   "api",
					Labels: map[string]string{"method": "Test", "location": "us-west-1b", "project_id": "test-project", "service": "test-service", "version": "v2"},
				},
				Severity:  loggingtypepb.LogSeverity_INFO,
				Timestamp: &timestamppb.Timestamp{Seconds: 1472728559},
				Operation: &loggingpb.LogEntryOperation{Id: "operation-id", Producer: "mixologist.io/mixologist"},
				Labels:    map[string]string{"http_method": "GET"},
				Payload:   &loggingpb.LogEntry_JsonPayload{JsonPayload: &structpb.Struct{Fields: map[string]*structpb.Value{"producer_project_id": &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "test-project"}}}}},
			},
		},
	}

	tl := &logger{}
	for _, v := range tests {
		f := &fakeClient{returnErr: v.reportErr}
		tl.c = f
		if err := tl.Log(v.in); err != nil && !v.reportErr {
			t.Errorf("%s: error received in Log(): %v", v.name, err)
		}

		if len(f.lastReq.Entries) != 1 {
			t.Errorf("%s: invalid number of log entries; got %d, want %d", v.name, len(f.lastReq.Entries), 1)
		}

		if !proto.Equal(f.lastReq.Entries[0], v.want) {
			t.Errorf("%s: bad request generated; got %v, want %v", v.name, f.lastReq.Entries[0], v.want)
		}
	}
}
