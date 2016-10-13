package config

const (
	// EveryService -- These rules apply to every service interaction
	EveryService = "_EVERY_SERVICE_"
)

type (
	// CacheParams -- configure caching behaviour
	// When adapter replies are cacheable, these params
	// configure behavious of the cache
	CacheParams struct {
	}

	// BatchParams -- configure batching behaviour of the adapter
	// These params should be used by the framework, *not* the adapter itself
	BatchParams struct {
		Size       int
		TimeoutSec int
	}
	// ConstructorParams -- 'Kind' is the adapter type
	// And Params are passed to the Kind constructor
	// This struct is sufficient to create a adapter
	ConstructorParams struct {
		// Kind - type of adapter
		Kind string
		// opaque params passed to the adapter
		Params interface{} `yaml:",omitempty"`
	}
	// AdapterParams -- constructor params and associated params
	AdapterParams struct {
		// Identifier of this adapter
		// should be unique within Kind, optional
		ID string
		// ConstructorParams embedded
		ConstructorParams `yaml:",omitempty,inline"`

		// cache params
		CacheParams CacheParams `yaml:",omitempty"`

		// batching params
		BatchParams BatchParams `yaml:",omitempty"`
	}

	// AdapterConfig -- in the given context
	// represents adapter configrations for all defined functions
	AdapterConfig struct {
		Checkers  []AdapterParams `yaml:",omitempty"`
		Reporters []AdapterParams `yaml:",omitempty"`
		// other functions
	}
	// BindingConfig -- same as adapter config + service
	// The other end of the service is understood by context
	BindingConfig struct {
		// ServiceID the service that this adapter config applies to
		ServiceID string
		// AdapterConfig -- embedded adapterconfig
		AdapterConfig `yaml:",omitempty,inline"`
	}

	// ServiceConfig -- Configuration from ServiceID's point of view
	ServiceConfig struct {
		// ServiceID the service that this adapter config applies to
		// this ServiceID also 'owns' the datastructure
		ServiceID string
		// Self -- configuration that applies to the service itself.
		// When service directly issues Check or Log / Report calls
		Self AdapterConfig `yaml:",omitempty"`
		// Ingress - applies When the given service is 'Producer'
		// irrespective of the calling entity
		Ingress AdapterConfig `yaml:",omitempty"`
		// Egress - applies When the given service is 'Consumer'
		// irrespective of the called entity
		Egress AdapterConfig `yaml:",omitempty"`
		// Consumers -- Applies to binding between binding.ServiceID --> ServiceID
		// binding.ServiceID is the consumer service of 'ServiceID'
		Consumers map[string]BindingConfig `yaml:",omitempty"`
		// Producers -- Applies to binding between ServiceID --> binding.ServiceID
		// binding.ServiceID is the Producer service for 'ServiceID'
		Producers map[string]BindingConfig `yaml:",omitempty"`
	}

	// ServicesConfig -- toplevel map describing all known services
	ServicesConfig map[string]ServiceConfig
)
