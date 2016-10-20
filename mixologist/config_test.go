package mixologist_test

import (
	"io/ioutil"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/cloudendpoints/mixologist/fakes"
	. "github.com/cloudendpoints/mixologist/mixologist"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

var (
	DirName string
)

func init() {
	_, filename, _, _ := runtime.Caller(1)
	DirName = path.Dir(filename)
}

func TestFormat(t *testing.T) {
	data, _ := ioutil.ReadFile(path.Join(DirName, "config/config_test.yml"))
	osc := ServicesConfig{}
	err := yaml.Unmarshal(data, &osc)
	if err != nil {
		t.Error("Unmarshal failed", err)
	}
	type RLParams struct {
		Rate string
	}

	for _, cl := range osc[InventoryService].Ingress.Checkers {
		if cl.Kind == "ratelimiter" {
			rlparam := RLParams{}
			mapstructure.Decode(cl.Params, &rlparam)
			var params map[interface{}]interface{}

			params = cl.Params.(map[interface{}]interface{})

			if params["rate"] != rlparam.Rate {
				t.Errorf("Expected: [ %v ]\nGot: [ %v ]", params["rate"], rlparam.Rate)
			}
		}
	}

	for _, cl := range osc[InventoryService].Self.Reporters {
		if cl.Kind == "statsd" {
			if cl.BatchParams.Size != 200 {
				t.Errorf("Expected: [ 200 ]\nGot: [ %v ]", cl.BatchParams)

			}
		}
	}
	if osc[InventoryService].Consumers[BindingID].Adapters.Checkers == nil {
		t.Errorf("Expected: binding not nil")
	}

	reg := map[string]CheckerBuilder{
		"fakechecker": fakes.NewCheckerBuilder("fakechecker", nil),
	}

	_, erra := ConvertParams(osc, reg)

	availabilityErrors := 0

	for _, er := range erra {
		if strings.HasSuffix(er.Error(), "is not available") {
			availabilityErrors += 1
		}
	}

	if (len(erra) - availabilityErrors) > 0 {
		t.Errorf("Got errors while converting %#v", erra[0])
	}
}

var (
	InventoryService = "Service.Inventory.1"
	ShippingService  = "Service.Shipping.1"
	BindingID        = "BindingID.1"
	SC               = ServicesConfig{
		EveryService: &ServiceConfig{
			ServiceID: EveryService,
			Ingress: &AdapterConfig{
				Reporters: []*AdapterParams{
					&AdapterParams{
						ConstructorParams: ConstructorParams{
							Kind: "gcloud.logging",
						},
					},
				},
			},
		},
		InventoryService: &ServiceConfig{
			ServiceID: InventoryService,
			Ingress: &AdapterConfig{
				Checkers: []*AdapterParams{
					&AdapterParams{
						ConstructorParams: ConstructorParams{
							Kind: "whitelist",
							Params: map[interface{}]interface{}{
								"provider_url": "http://mywhitelist",
							},
						}},
					&AdapterParams{
						ConstructorParams: ConstructorParams{
							// By default this service allows
							// 100 req /s
							Kind: "ratelimiter",
							Params: map[interface{}]interface{}{
								"rate": "100/s",
							},
						}},
				},
				Reporters: []*AdapterParams{
					&AdapterParams{
						ConstructorParams: ConstructorParams{
							Kind: "statsd",
							Params: map[interface{}]interface{}{
								"host": "statsd",
								"port": 9317,
							},
						}},
				},
			},
			Consumers: map[string]*BindingConfig{
				BindingID: &BindingConfig{
					ShippingService,
					&AdapterConfig{
						Checkers: []*AdapterParams{
							&AdapterParams{
								ConstructorParams: ConstructorParams{
									// For Shipping Service, this service allows
									// a higher rate
									Kind: "ratelimiter",
									Params: map[interface{}]interface{}{
										"rate": "1000/s",
									},
								},
							}},
					},
				},
			},
		},
		ShippingService: &ServiceConfig{
			ServiceID: ShippingService,
			// Send my logs to aws.logging regardless of who I am calling
			Egress: &AdapterConfig{
				Reporters: []*AdapterParams{
					&AdapterParams{
						ConstructorParams: ConstructorParams{
							Kind: "aws.logging",
						},
					}},
			},
			Producers: map[string]*BindingConfig{
				BindingID: &BindingConfig{
					InventoryService,
					&AdapterConfig{
						Checkers: []*AdapterParams{
							&AdapterParams{
								ConstructorParams: ConstructorParams{
									// Inventory Service is expensive to call
									// I (ShippingService) wants to impose a lower limit
									Kind: "ratelimiter",
									Params: map[interface{}]interface{}{
										"rate": "5/s",
									},
								},
							}},
					},
				},
			},
		},
	}
)
