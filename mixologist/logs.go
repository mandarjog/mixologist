package mixologist

import (
	"bytes"
	"encoding/json"
	"time"
)

const (
	DefaultLogger = "mixologist.io/loggers/default"
)

type (
	// Logger provides a basic interface for writing directly to
	// a logs sink.
	Logger interface {
		Name() string
		Log(LogEntry) error
		Flush()
	}

	// LoggerBuilders build Loggers for use as logs sinks
	LoggerBuilder interface {
		// Given system config, build a new logger
		Build(Config) Logger
	}

	// Resource represents a monitored resource in mixologist.
	// For now, the only supported type of monitored resource is API.
	Resource struct {
		// type. should always be "api".
		Type   string            `json:"type,omitempty"`
		Labels map[string]string `json:"labels,omitempty"`
	}

	// LogEntry is the top-level struct for logs data that will be
	// exported by this adapter. It is used for more digestible
	// json-formatted output from the exposed format in the mixologist
	// API.
	LogEntry struct {
		Name          string                 `json:"logName,omitempty"`
		Timestamp     time.Time              `json:"timestamp,omitempty"`
		OperationID   string                 `json:"operationId,omitempty"`
		ID            string                 `json:"id, omitempty"`
		Resource      Resource               `json:"resource,omitempty"`
		Labels        map[string]string      `json:"labels,omitempty"`
		Severity      string                 `json:"severity,omitempty"`
		ProtoPayload  string                 `json:"protoPayload,omitempty"`
		TextPayload   string                 `json:"textPayload,omitempty"`
		StructPayload map[string]interface{} `json:"structPayload,omitempty"`
	}
)

func JSONBytes(l LogEntry) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(l); err != nil {
		// TODO(dougreid): do we want to err out here?
		return []byte{}, err
	}
	return buf.Bytes(), nil
}
