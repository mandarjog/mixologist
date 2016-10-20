package main

import (
	"bytes"
	"flag"
	"fmt"
	sc "google/api/servicecontrol/v1"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cloudendpoints/mixologist/testutils"
	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
)

var (
	serviceName   = flag.String("dest", "service1", "service")
	consumerID    = flag.String("source", "", "source")
	operationName = flag.String("oprn", "", "Operation Name")
	operationId   = flag.String("oprn-id", "TEST_OPRN_ID", "Operation Id")
	callerIp      = flag.String("callerip", "127.0.0.1", "Caller IP")
	mixologist    = flag.String("mixologist", "http://localhost:9092", "mixologist url")
)

func main() {
	flag.Parse()
	cr := testutils.CreateCheck(&testutils.ExpectedCheck{
		ServiceName:   *serviceName,
		ConsumerID:    *consumerID,
		OperationName: *operationName,
		OperationId:   *operationId,
		CallerIp:      *callerIp,
	})
	url := fmt.Sprintf("%s/v1/services/%s:check", *mixologist, *serviceName)
	data, err := proto.Marshal(&cr)
	if err != nil {
		glog.Exitf("Unable to marshal %#v", err)
	}
	clnt := &http.Client{
		Timeout: time.Second * 5,
	}

	resp, err := clnt.Post(url, "protobuf", bytes.NewReader(data))
	if err != nil {
		glog.Exitf("Unable to POST %v", err)
	}

	bdata, err := ioutil.ReadAll(resp.Body)
	glog.Infof("%#v", resp.Header)
	cresp := sc.CheckResponse{}
	proto.Unmarshal(bdata, &cresp)
	if len(cresp.GetCheckErrors()) > 0 {
		glog.Info("CheckError %#v", cresp)
	} else {
		glog.Info("Check OK")
	}
}
