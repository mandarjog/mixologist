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
func NewHandler(server ServiceControllerServer, hh []*PrefixAndHandler, opts ...func(*Handler)) *Handler {
	h := &Handler{
		Server:         server,
		ReportHandlers: hh,
		readf:          ioutil.ReadAll,
		marshal:        proto.Marshal,
		unmarshal:      proto.Unmarshal,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// Marshal -- specify an override fn for marshaling protobuf default: proto.Marshal
func Marshal(marshal marshalfn) func(*Handler) {
	return func(h *Handler) {
		h.marshal = marshal
	}
}

// ReadHTTPBody -- provide alternate implementation for reading http body. default: ioutil.ReadAll
func ReadHTTPBody(readf readfn) func(*Handler) {
	return func(h *Handler) {
		h.readf = readf
	}
}

// Perform common preamble during message specific processing
func (h *Handler) preambleProcess(w http.ResponseWriter, r *http.Request, msg proto.Message) (err error) {
	body, err := h.readf(r.Body)
	if err != nil {
		return
	}
	err = h.unmarshal(body, msg)
	if err != nil {
		return
	}

	glog.V(2).Infoln(msg.String())
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

	if strings.HasSuffix(r.RequestURI, CheckSuffix) {
		return h.serverCheck
	} else if strings.HasSuffix(r.RequestURI, ReportSuffix) {
		return h.serverReport
	}

	glog.Warning("Got unknown URI " + r.RequestURI)
	w.WriteHeader(http.StatusNotFound)
	return nil
}

// Implement Handler API
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// check for registered handlers
	for _, ph := range h.ReportHandlers {
		if strings.HasPrefix(r.RequestURI, ph.Prefix) {
			ph.Handler.ServeHTTP(w, r)
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
		w.Write([]byte(err.Error()))
		glog.Error(err)
		return
	}
	if respb, err := h.marshal(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		glog.Error(err)
	} else {
		w.Write(respb)
	}
}
