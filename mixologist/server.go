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

// NewHandler -- return a handler with initialized handler map
func NewHandler(server ServiceControllerServer) *Handler {
	return &Handler{
		Server:     server,
		HandlerMap: make(map[string]http.Handler),
	}

}

// AddHandler -- Add a handler for prefixes, unsed for services like
// prometheus
func (h *Handler) AddHandler(prefix string, hh http.Handler) {
	h.HandlerMap[prefix] = hh
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

// ServerCheck -- wrapper for Server.Check
func (h *Handler) serverCheck(w http.ResponseWriter, r *http.Request, ctx context.Context) (resp proto.Message, err error) {
	msg := &sc.CheckRequest{}

	if err = h.preambleProcess(w, r, msg); err != nil {
		return nil, err
	}
	return h.Server.Check(ctx, msg)
}

// ServerReport -- wrapper for Server.Report
func (h *Handler) serverReport(w http.ResponseWriter, r *http.Request, ctx context.Context) (resp proto.Message, err error) {
	msg := &sc.ReportRequest{}

	if err = h.preambleProcess(w, r, msg); err != nil {
		return nil, err
	}
	return h.Server.Report(ctx, msg)
}

type serverFn func(w http.ResponseWriter, r *http.Request, ctx context.Context) (resp proto.Message, err error)

// GetServerFn -- Find the handler server FN - Returns nil if unable to find
func (h *Handler) getServerFn(w http.ResponseWriter, r *http.Request) serverFn {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}

	if strings.HasSuffix(r.RequestURI, ":check") {
		return h.serverCheck
	} else if strings.HasSuffix(r.RequestURI, ":report") {
		return h.serverReport
	}

	glog.Warning("Got unknown URI " + r.RequestURI)
	w.WriteHeader(http.StatusNotFound)
	return nil
}

// Implement Handler API
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// check for registered handlers
	for prefix, hh := range h.HandlerMap {
		if strings.HasPrefix(r.RequestURI, prefix) {
			hh.ServeHTTP(w, r)
			return
		}
	}

	var fn = h.getServerFn(w, r)
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
