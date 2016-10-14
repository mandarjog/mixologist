package mixologist_test

import (
	"github.com/cloudendpoints/mixologist/fakes"
	. "github.com/cloudendpoints/mixologist/mixologist"
	"github.com/cloudendpoints/mixologist/mixologist/config"
	g "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestCheckerManager(t *testing.T) {
	g.RegisterTestingT(t)

	// empty test
	cfg := config.ServicesConfig{}
	reg := map[string]CheckerBuilder{
		"fakechecker": fakes.NewCheckerBuilder("fakechecker", nil),
	}
	cm, erra := NewCheckerManager(reg, cfg)
	g.Expect(cm).NotTo(g.BeNil())
	g.Expect(erra).To(g.BeEmpty())
	g.Expect(cm.Checkers()).Should(g.BeEmpty())
}

var yamlStr = `
_EVERY_SERVICE_:  # Applies to every interaction:- between all consumers and producers
  serviceid: _EVERY_SERVICE_
  ingress:
    reporters:
    - kind: statsd
      params:
          addr: "statsd:8125"
    - kind: prometheus
    - kind: mixologist.io/consumers/logsAdapter
      params:
          backends:
              - glog
              - stackdriver
    checkers:
`
var whitelist = `
    - kind: whitelist
      params:
          providerurl: https://gist.githubusercontent.com/mandarjog/c38f4a992cc5d470ad763e70eca709b9/raw/
`
var fakechecker = `
    - kind: fakechecker
      params:
          oncall: supercoder@acme
          flist:
                wl: abcdefg
`
var fakecheckerMissingRequired = `
    - kind: fakechecker
      params:
          oncall: supercoder@acme
`
var fakecheckerWrongType = `
    - kind: fakechecker
      params:
          oncall: supercoder@acme
          flist:
                wl: 2000
`

func TestCheckerManagerValidate(t *testing.T) {
	g.RegisterTestingT(t)
	cfg := config.ServicesConfig{}
	yaml.Unmarshal([]byte(yamlStr+whitelist), &cfg)
	g.Expect(cfg[config.EveryService].Ingress.Checkers).ShouldNot(g.BeEmpty())
	reg := map[string]CheckerBuilder{
		"fakechecker": fakes.NewCheckerBuilder("fakechecker", nil),
	}
	cm, erra := NewCheckerManager(reg, cfg)
	g.Expect(cm).NotTo(g.BeNil())
	g.Expect(erra).NotTo(g.BeEmpty())
	ers := ErrAdapterUnavailable("whitelist")
	g.Expect(erra[0]).To(g.Equal(ers))
	g.Expect(cm.Checkers()).Should(g.BeEmpty())
}

func TestCheckerManagerValidate1(t *testing.T) {
	g.RegisterTestingT(t)
	cfg := config.ServicesConfig{}
	yaml.Unmarshal([]byte(yamlStr+fakechecker), &cfg)
	g.Expect(cfg[config.EveryService].Ingress.Checkers).ShouldNot(g.BeEmpty())
	reg := map[string]CheckerBuilder{
		"fakechecker": fakes.NewCheckerBuilder("fakechecker", nil),
	}
	cm, erra := NewCheckerManager(reg, cfg)
	g.Expect(cm).ShouldNot(g.BeNil())
	g.Expect(erra).Should(g.BeEmpty())
	g.Expect(cm.Checkers()).ShouldNot(g.BeEmpty())
}

func TestCheckerManagerValidate2(t *testing.T) {
	g.RegisterTestingT(t)
	cfg := config.ServicesConfig{}
	yaml.Unmarshal([]byte(yamlStr+fakecheckerMissingRequired), &cfg)
	g.Expect(cfg[config.EveryService].Ingress.Checkers).ShouldNot(g.BeEmpty())
	reg := map[string]CheckerBuilder{
		"fakechecker": fakes.NewCheckerBuilder("fakechecker", nil),
	}
	cm, erra := NewCheckerManager(reg, cfg)
	g.Expect(cm).ShouldNot(g.BeNil())
	g.Expect(erra).ShouldNot(g.BeEmpty())
	ve := erra[0].(*DecodeError)
	g.Expect(ve.Missing[0]).To(g.Equal("Flist.Wl"))
	g.Expect(cm.Checkers()).Should(g.BeEmpty())
}

func TestCheckerManagerValidate3(t *testing.T) {
	g.RegisterTestingT(t)
	cfg := config.ServicesConfig{}
	yaml.Unmarshal([]byte(yamlStr+fakecheckerWrongType), &cfg)
	g.Expect(cfg[config.EveryService].Ingress.Checkers).ShouldNot(g.BeEmpty())
	reg := map[string]CheckerBuilder{
		"fakechecker": fakes.NewCheckerBuilder("fakechecker", nil),
	}
	cm, erra := NewCheckerManager(reg, cfg)
	g.Expect(cm).ShouldNot(g.BeNil())
	g.Expect(erra).ShouldNot(g.BeEmpty())
	ve := erra[0].(*DecodeError)
	g.Expect(ve.Error()).To(g.ContainSubstring("unconvertible type"))
	g.Expect(cm.Checkers()).Should(g.BeEmpty())
}