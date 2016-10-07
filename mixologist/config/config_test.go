package config

import (
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
	"runtime"
	"testing"
)

var (
	DirName string
)

func init() {
	_, filename, _, _ := runtime.Caller(1)
	DirName = path.Dir(filename)
}

func TestFormat(t *testing.T) {
	data, _ := ioutil.ReadFile(path.Join(DirName, "config_test.yml"))
	osc := ServicesConfig{}
	yaml.Unmarshal(data, &osc)

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

}

var (
	InventoryService = "Service.Inventory.1"
	ShippingService  = "Service.Shipping.1"
	BindingID        = "BindingID.1"
	SC               = ServicesConfig{
		EveryService: ServiceConfig{
			ServiceID: EveryService,
			Ingress: AdapterConfig{
				Reporters: []ConstructorParams{
					ConstructorParams{
						Kind: "gcloud.logging",
					},
				},
			},
		},
		InventoryService: ServiceConfig{
			ServiceID: InventoryService,
			Ingress: AdapterConfig{
				Checkers: []ConstructorParams{
					ConstructorParams{
						Kind: "whitelist",
						Params: map[string]interface{}{
							"provider_url": "http://mywhitelist",
						},
					},
					ConstructorParams{
						// By default this service allows
						// 100 req /s
						Kind: "ratelimiter",
						Params: map[string]interface{}{
							"rate": "100/s",
						},
					},
				},
				Reporters: []ConstructorParams{
					ConstructorParams{
						Kind: "statsd",
						Params: map[string]interface{}{
							"host": "statsd",
							"port": 9317,
						},
					},
				},
			},
			Consumers: map[string]BindingConfig{
				BindingID: BindingConfig{
					ShippingService,
					AdapterConfig{
						Checkers: []ConstructorParams{
							ConstructorParams{
								// For Shipping Service, this service allows
								// a higher rate
								Kind: "ratelimiter",
								Params: map[string]interface{}{
									"rate": "1000/s",
								},
							},
						},
					},
				},
			},
		},
		ShippingService: ServiceConfig{
			ServiceID: ShippingService,
			// Send my logs to aws.logging regardless of who I am calling
			Egress: AdapterConfig{
				Reporters: []ConstructorParams{
					ConstructorParams{
						Kind: "aws.logging",
					},
				},
			},
			Producers: map[string]BindingConfig{
				BindingID: BindingConfig{
					InventoryService,
					AdapterConfig{
						Checkers: []ConstructorParams{
							ConstructorParams{
								// Inventory Service is expensive to call
								// I (ShippingService) wants to impose a lower limit
								Kind: "ratelimiter",
								Params: map[string]interface{}{
									"rate": "5/s",
								},
							},
						},
					},
				},
			},
		},
	}
)
