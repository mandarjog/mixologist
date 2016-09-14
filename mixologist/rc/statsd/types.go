package statsd

import (
	sd "github.com/cactus/go-statsd-client/statsd"
	"somnacin-internal/mixologist/mixologist"
)

func init() {
	mixologist.RegisterReportConsumer(Name, new(builder))
}

type (
	consumer struct {
		client sd.Statter
	}
	builder struct{}
)
