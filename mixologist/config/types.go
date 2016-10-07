package config

const (
	// EveryService -- These rules apply to every service interaction
	EveryService = "_EVERY_SERVICE_"
)

type (
	// ConstructorParams -- 'Kind' is the adapter type
	// And Params are passed to the Kind constructor
	// This struct is sufficient to create a adapter
	ConstructorParams struct {
		// Kind - type of adapter
		Kind string
		// opaque params passed to the adapter
		Params interface{} ",omitempty"
	}

	// AdapterConfig -- in the given context
	// represents adapter configrations for all defined functions
	AdapterConfig struct {
		Checkers  []ConstructorParams ",omitempty"
		Reporters []ConstructorParams ",omitempty"
		// other functions
	}
	// BindingConfig -- same as adapter config + service
	// The other end of the service is understood by context
	BindingConfig struct {
		// ServiceID the service that this adapter config applies to
		ServiceID string
		// AdapterConfig -- embedded adapterconfig
		AdapterConfig ",omitempty,inline"
	}

	ServiceConfig struct {
		// ServiceID the service that this adapter config applies to
		// this ServiceID also 'owns' the datastructure
		ServiceID string
		// Self -- configuration that applies to the service itself.
		// When service directly issues Check or Log / Report calls
		Self AdapterConfig ",omitempty"
		// Ingress - applies When the given service is 'Producer'
		// irrespective of the calling entity
		Ingress AdapterConfig ",omitempty"
		// Egress - applies When the given service is 'Consumer'
		// irrespective of the called entity
		Egress AdapterConfig ",omitempty"
		// Consumers -- Applies to binding between binding.ServiceID --> ServiceID
		// binding.ServiceID is the consumer service of 'ServiceID'
		Consumers map[string]BindingConfig ",omitempty"
		// Producers -- Applies to binding between ServiceID --> binding.ServiceID
		// binding.ServiceID is the Producer service for 'ServiceID'
		Producers map[string]BindingConfig ",omitempty"
	}

	ServicesConfig map[string]ServiceConfig
)
