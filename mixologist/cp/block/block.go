package block

import (
	sc "google/api/servicecontrol/v1"

	"github.com/cloudendpoints/mixologist/mixologist"
)

const (
	Name           = "block"
	DefaultMessage = "Access explicitly blocked"
)

type (
	builder struct{}
	checker struct {
		cfg *Config
	}
	Config struct {
		Message string `yaml:"message"`
	}
)

func checkError(message string) *sc.CheckError {
	return &sc.CheckError{
		Code:   sc.CheckError_CLIENT_APP_BLOCKED,
		Detail: message,
	}
}

func init() {
	mixologist.RegisterChecker(Name, new(builder))
}

func (c *checker) Name() string {
	return Name
}

func (c *checker) Unload() {}

func (c *checker) Check(cr *sc.CheckRequest) (*sc.CheckError, error) {
	return checkError(c.cfg.Message), nil
}

// BuildChecker -- exported method
func (b *builder) BuildChecker(cfg interface{}) (mixologist.Checker, error) {
	cc := cfg.(*Config)
	return &checker{
		cfg: cc,
	}, nil
}

// ConfigStruct -- return pointer to Config struct
func (b *builder) ConfigStruct() interface{} {
	return &Config{
		Message: DefaultMessage,
	}
}

// ValidateConfig -- validate given config
func (b *builder) ValidateConfig(cfg interface{}) error { return nil }
