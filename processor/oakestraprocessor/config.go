package oakestraprocessor // import github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor

import (
	"go.opentelemetry.io/collector/component"
)

var _ component.ConfigValidator = (*Config)(nil)

// Config represents the processor config settings within the collector's config.yaml
type Config struct {
}

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	return nil
}
