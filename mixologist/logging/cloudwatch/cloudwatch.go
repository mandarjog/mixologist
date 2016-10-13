// Package cloudwatch enables logging to AWS cloudwatchlogs.
//
// To use this package, one must currently create a credentials.ini file that contains
// the looks like the following:
//
// [default]
// aws_access_key_id=<KEY>
// aws_secret_access_key=<SECRET>
//
// Then, one must create a kubernetes secret in the namespace that mixologist will be
// run in.  Then the mixologist RC must be modified to mount the volume to /etc/aws.
//
// In the future, there will likely be a better, more dynamic way of passing in creds for
// use in this package.
package cloudwatch

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/cloudendpoints/mixologist/mixologist"
	"github.com/cloudendpoints/mixologist/mixologist/rc/logsAdapter"
	"github.com/golang/glog"
)

const (
	// Name is the unique identifier for this (logs) adapter
	Name = "aws/cloudwatchlogs"

	// Region selects the AWS region (and endpoint) to send logs.
	// TODO: make a configuration option
	Region = "us-west-2" // US West (Oregon) us-west-2 logs.us-west-2.amazonaws.com HTTPS

	// creds details (const for now -- could move into supplier)
	svcAcctFile    = "/etc/aws/credentials.ini"
	svcAcctProfile = "default"
)

func init() { logsAdapter.RegisterLogsSink(Name, new(builder)) }

type (
	logger struct {
		c cloudwatchlogsiface.CloudWatchLogsAPI

		// TODO: log streams must be published to with sequence tokens.
		// as mixologist scales, care must be taken to handle multiple
		// logs producers writing to the same the log stream (within the
		// same log group and log region)
		//
		// One possibility: each individual mixologist instance could
		// uniquely identify itself and create a unique log stream, within
		// a log group (perhaps log group per service? (limit 500 groups per account))
		//
		// TODO: also, this should probably be synchronized in some fashion
		seqTokens map[string]string
	}

	builder struct{}
)

func (b *builder) Build(c mixologist.Config) mixologist.Logger {

	creds := credentials.NewSharedCredentials(svcAcctFile, svcAcctProfile)
	config := aws.NewConfig().WithCredentials(creds).WithRegion(Region)
	sess, err := session.NewSession(config)
	if err != nil {
		glog.Errorf("could not create aws cloudwatchlogs session: %v", err)
		return nil
	}
	svc := cloudwatchlogs.New(sess)
	return &logger{c: svc, seqTokens: make(map[string]string)}
}

func (l *logger) Name() string { return Name }
func (l *logger) Flush()       {}
func (l *logger) Log(le mixologist.LogEntry) error {

	glog.V(2).Infof("logging to aws cloudwatchlogs: %v", le)

	// TODO: decide on log group and log stream naming
	group := le.Name
	stream := serviceName(le)

	seq, err := l.setupLogs(group, stream)
	if err != nil {
		glog.Errorf("could not retrieve sequence token: %v", err)
		return err
	}

	req := &cloudwatchlogs.PutLogEventsInput{
		LogEvents:     []*cloudwatchlogs.InputLogEvent{convert(le)},
		LogGroupName:  aws.String(group),  // Required
		LogStreamName: aws.String(stream), // Required
	}
	if seq != "" { // covers new stream case
		req.SequenceToken = aws.String(seq)
	}

	glog.V(2).Infof("aws cloudwatchlogs request: %v", req)

	// TODO: might need to retry on seq token errors in multi-producer cases
	resp, err := l.c.PutLogEvents(req)
	if err != nil {
		glog.Errorf("could not log to aws: %v", err)
		return err
	}

	// TODO: punting on sync for now around seq tokens
	l.seqTokens[stream] = aws.StringValue(resp.NextSequenceToken)
	glog.V(2).Infof("next seq token: %v", l.seqTokens[stream])

	return nil
}

// TODO: probably a better way to do upon adapter initialization
func (l *logger) setupLogs(group, stream string) (string, error) {

	// log group & stream exist and have been previously used
	if val, ok := l.seqTokens[stream]; ok {
		return val, nil
	}

	err := l.createGroup(group)
	if err != nil {
		return "", err
	}

	err = l.createStream(group, stream)
	if err != nil {
		return "", err
	}

	token, err := l.getSequenceToken(group, stream)
	if err != nil {
		return "", err
	}

	// TODO: punting on sync for now around seq tokens
	l.seqTokens[stream] = token
	return token, nil
}

func (l *logger) createGroup(group string) error {
	glog.V(2).Infof("creating aws log group: %s", group)
	params := &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(group), // Required
	}
	_, err := l.c.CreateLogGroup(params)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "ResourceAlreadyExistsException":
				return nil
			}
		}
		glog.Errorf("could not create aws log group '%s': %v", group, err)
		return err
	}

	return nil
}

func (l *logger) createStream(group, stream string) error {
	glog.V(2).Infof("creating aws log stream '%s' in group '%s'", stream, group)

	params := &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(group),  // Required
		LogStreamName: aws.String(stream), // Required
	}
	_, err := l.c.CreateLogStream(params)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "ResourceAlreadyExistsException":
				return nil
			}
		}
		glog.Errorf("could not create aws log stream '%s': %v", stream, err)
		return err
	}

	return nil
}

func (l *logger) getSequenceToken(group, stream string) (string, error) {
	glog.V(2).Infof("getting aws log stream data for '%s'", stream)

	params := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        aws.String(group), // Required
		LogStreamNamePrefix: aws.String(stream),
		Limit:               aws.Int64(1),
	}
	resp, err := l.c.DescribeLogStreams(params)
	if err != nil {
		glog.Errorf("could not get stream data '%s': %v", stream, err)
		return "", err
	}

	for _, ls := range resp.LogStreams {
		if aws.StringValue(ls.LogStreamName) == stream {
			glog.Infof("found stream; %v", ls)
			return aws.StringValue(ls.UploadSequenceToken), nil
		}
	}

	return "", errors.New("no token found for streams")
}

// TODO: this probably needs to be more robust
func serviceName(le mixologist.LogEntry) string {
	name := le.Labels[mixologist.CloudService]
	if name == "" {
		name = le.Resource.Labels[mixologist.CloudService]
	}
	return name
}

// TODO: add size checks, etc.
func convert(l mixologist.LogEntry) *cloudwatchlogs.InputLogEvent {

	t := int64((time.Duration(l.Timestamp.UnixNano()) * time.Nanosecond) / time.Millisecond)
	message, err := json.Marshal(l)
	if err != nil {
		glog.Errorf("error marshalling json: %v", err)
		return nil
	}

	le := &cloudwatchlogs.InputLogEvent{
		Message:   aws.String(string(message)),
		Timestamp: aws.Int64(t),
	}

	return le
}
