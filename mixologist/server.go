package mixologist

import (
	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	sc "google/api/servicecontrol/v1"
	"io/ioutil"
	"net/http"
	"strings"
)

func NewHandler(server ServiceControllerServer) *Handler {
	return &Handler{
		Server: server,
	}

}

// Perform common preamble during message specific processing
func (h *Handler) preambleProcess(w http.ResponseWriter, r *http.Request, msg proto.Message) (err error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		glog.Error(err)
		return
	}
	err = proto.Unmarshal(body, msg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		glog.Error(err)
		return
	}

	if glog.V(2) {
		glog.Info(msg.String())
	}
	return
}

// wrapper for Server.Check
func (h *Handler) ServerCheck(w http.ResponseWriter, r *http.Request, ctx context.Context) (resp proto.Message, err error) {
	msg := &sc.CheckRequest{}

	if err = h.preambleProcess(w, r, msg); err != nil {
		return nil, err
	}
	return h.Server.Check(ctx, msg)
}

// wrapper for Server.Report
func (h *Handler) ServerReport(w http.ResponseWriter, r *http.Request, ctx context.Context) (resp proto.Message, err error) {
	msg := &sc.ReportRequest{}

	if err = h.preambleProcess(w, r, msg); err != nil {
		return nil, err
	}
	return h.Server.Report(ctx, msg)
}

type ServerFn func(w http.ResponseWriter, r *http.Request, ctx context.Context) (resp proto.Message, err error)

// Find the handler server FN - Returns nil if unable to find
func (h *Handler) GetServerFn(w http.ResponseWriter, r *http.Request) ServerFn {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}

	if strings.HasSuffix(r.RequestURI, ":check") {
		return h.ServerCheck
	} else if strings.HasSuffix(r.RequestURI, ":report") {
		return h.ServerReport
	}

	glog.Warning("Got unknown URI " + r.RequestURI)
	w.WriteHeader(http.StatusNotFound)
	return nil
}

// Implement Handler API
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var fn = h.GetServerFn(w, r)
	if fn == nil {
		return
	}

	ctx := context.Background()

	resp, err := fn(w, r, ctx)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		glog.Error(err)
		return
	}

	respb, err := proto.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		glog.Error(err)
		return
	}
	w.Write(respb)
}
