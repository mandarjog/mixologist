package cloudwatch

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/cloudendpoints/mixologist/mixologist"
)

type fakeClient struct {
	cloudwatchlogsiface.CloudWatchLogsAPI

	returnErr bool
	lastReq   *cloudwatchlogs.PutLogEventsInput
}

func (f *fakeClient) CreateLogGroup(req *cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error) {
	return nil, nil
}

func (f *fakeClient) CreateLogStream(req *cloudwatchlogs.CreateLogStreamInput) (*cloudwatchlogs.CreateLogStreamOutput, error) {
	return nil, nil
}

func (f *fakeClient) DescribeLogStreams(req *cloudwatchlogs.DescribeLogStreamsInput) (*cloudwatchlogs.DescribeLogStreamsOutput, error) {
	return &cloudwatchlogs.DescribeLogStreamsOutput{
		LogStreams: []*cloudwatchlogs.LogStream{
			&cloudwatchlogs.LogStream{
				LogStreamName:       req.LogStreamNamePrefix,
				UploadSequenceToken: aws.String("token"),
			},
		},
	}, nil
}

func (f *fakeClient) PutLogEvents(req *cloudwatchlogs.PutLogEventsInput) (*cloudwatchlogs.PutLogEventsOutput, error) {
	if f.returnErr {
		f.lastReq = nil
		return nil, errors.New("error writing logs")
	}
	f.lastReq = req
	return &cloudwatchlogs.PutLogEventsOutput{NextSequenceToken: aws.String("next_token")}, nil
}

func TestLog(t *testing.T) {
	tests := []struct {
		name      string
		in        mixologist.LogEntry
		want      *cloudwatchlogs.PutLogEventsInput
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
			want: &cloudwatchlogs.PutLogEventsInput{
				LogGroupName:  aws.String("test-log"),
				LogStreamName: aws.String("test-service"),
				SequenceToken: aws.String("token"),
				LogEvents: []*cloudwatchlogs.InputLogEvent{
					&cloudwatchlogs.InputLogEvent{
						Message:   aws.String("{\"logName\":\"test-log\",\"timestamp\":\"2016-09-01T11:15:59Z\",\"operationId\":\"operation-id\",\"id\":\"\",\"resource\":{\"type\":\"api\",\"labels\":{\"cloud.googleapis.com/location\":\"us-west-1b\",\"cloud.googleapis.com/project\":\"test-project\",\"cloud.googleapis.com/service\":\"test-service\",\"serviceruntime.googleapis.com/api_method\":\"Test\",\"serviceruntime.googleapis.com/api_version\":\"v2\"}},\"labels\":{\"http_method\":\"GET\"},\"severity\":\"INFO\",\"structPayload\":{\"producer_project_id\":\"test-project\"}}"),
						Timestamp: aws.Int64(1472728559000),
					},
				},
			},
		},
	}

	tl := &logger{seqTokens: make(map[string]string)}
	for _, v := range tests {
		f := &fakeClient{returnErr: v.reportErr}
		tl.c = f
		if err := tl.Log(v.in); err != nil && !v.reportErr {
			t.Errorf("%s: error received in Log(): %v", v.name, err)
		}

		// if len(f.lastReq.Entries) != 1 {
		// 	t.Errorf("%s: invalid number of log entries; got %d, want %d", v.name, len(f.lastReq.Entries), 1)
		// }

		if !reflect.DeepEqual(f.lastReq, v.want) {
			t.Errorf("%s: bad request generated; got %v, want %v", v.name, f.lastReq, v.want)
		}
	}
}
