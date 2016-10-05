package glog

import (
	"github.com/golang/glog"
	"github.com/cloudendpoints/mixologist/mixologist"
	"github.com/cloudendpoints/mixologist/mixologist/rc/logsAdapter"
)

const Name = "mixologist.io/loggers/glog"

func init() { logsAdapter.RegisterLogsSink(Name, new(builder)) }

type (
	logger  struct{}
	builder struct{}
)

func (b *builder) Build(c mixologist.Config) mixologist.Logger { return &logger{} }

func (l *logger) Name() string { return Name }
func (l *logger) Log(le mixologist.LogEntry) error {
	out, err := mixologist.JSONBytes(le)
	if err != nil {
		return err
	}
	glog.Infof("%s", out)
	return nil
}
func (l *logger) Flush() { glog.Flush() }
