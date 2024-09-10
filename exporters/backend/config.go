package backendexporter // import github.com/smnzlnsk/opentelemetry-components/exporters/backend

// Config represents the receiver config settings within the collector's config.yaml
type Config struct{}

// Validate checks if the receiver configuration is valid
func (cfg *Config) Validate() error {
	return nil
}
