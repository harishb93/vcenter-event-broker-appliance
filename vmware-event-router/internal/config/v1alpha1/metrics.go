package v1alpha1

// MetricsProviderType represents a supported metrics provider
type MetricsProviderType string

const (
	// MetricsProviderDefault is the the default metrics provider
	MetricsProviderDefault    MetricsProviderType = "default"
	MetricsProviderPrometheus MetricsProviderType = "prometheus"
)

// MetricsProvider configures the metrics provider
type MetricsProvider struct {
	// Type sets the metrics provider
	Type MetricsProviderType `yaml:"type" json:"type" jsonschema:"enum=default,enum=prometheus,required"`
	// Name is an identifier for the configured metrics provider
	Name string `yaml:"name" json:"name" jsonschema:"required"`
	// +optional
	Default *MetricsProviderConfigDefault `yaml:"default,omitempty" json:"default,omitempty" jsonschema:"oneof_required=default"`
	// Prometheus Config - Required if type 'prometheus'
	Prometheus *MetricsProviderConfigPrometheus `yaml:"prometheus,omitempty" json:"prometheus,omitempty" jsonschema:"oneof_required=prometheus"`
}

// MetricsProviderConfigDefault configures the default metrics provider
type MetricsProviderConfigDefault struct {
	// BindAddress is the address where the default metrics provider http endpoint will listen for connections
	BindAddress string `yaml:"bindAddress" json:"bindAddress" jsonschema:"required,default=0.0.0.0:8082"`
	// Auth when specified requires authentication for the http endpoint of the
	// metrics provider. Only basic_auth is supported.
	// +optional
	Auth *AuthMethod `yaml:"auth,omitempty" json:"auth,omitempty" jsonschema:"description=Authentication configuration for this section"`
}

type MetricsProviderConfigPrometheus struct {
	//Placeholder
	// Address is the address where the prometheus metrics provider http endpoint will be running
	BindAddress string `yaml:"bindAddress" json:"bindAddress" jsonschema:"required,default=0.0.0.0:9090"`
}
