package statsd

import (
	sd "github.com/cactus/go-statsd-client/statsd"
	"somnacin-internal/mixologist/mixologist"
)

func init() {
	mixologist.RegisterReportConsumer(Name, new(builder))
}

type (
	// ServerConfig contains configuration info for a statsd backend
	ServerConfig struct {
		Addr string
	}
	consumer struct {
		client sd.Statter
	}
	builder struct{}
)
