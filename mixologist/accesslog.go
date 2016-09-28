package mixologist

import (
	"log"
	"net/http"
	"time"
)

type loggingWriter struct {
	w      http.ResponseWriter
	status int
	size   int
}

func (s *loggingWriter) Header() http.Header { return s.w.Header() }

func (s *loggingWriter) Write(b []byte) (int, error) {
	size, err := s.w.Write(b)
	s.size += size
	return size, err
}

func (s *loggingWriter) WriteHeader(status int) {
	s.w.WriteHeader(status)
	s.status = status
}

func (s loggingWriter) Flush() {
	if flusher, found := s.w.(http.Flusher); found {
		flusher.Flush()
	}
}

func (s loggingWriter) LogAccess(r *http.Request, tt time.Duration) {
	log.Printf("AccessLog: %d %s %s %s %d %d %s", s.status, r.Method, r.RequestURI, r.RemoteAddr, r.ContentLength, s.size, tt.String())
}

// BuildLoggingWriter -- create logging writer that retains status and size
func buildLoggingWriter(w http.ResponseWriter) *loggingWriter {
	return &loggingWriter{
		w:      w,
		status: http.StatusOK,
		size:   0,
	}
}
