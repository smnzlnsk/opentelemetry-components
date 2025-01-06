package oakestraheuristicengine

import (
	"go.opentelemetry.io/collector/component"
)

// Config defines the configuration for the oakestraheuristicengine processor.
type Config struct {
	// Add your configuration fields here
}

var _ component.Config = (*Config)(nil)

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	return nil
}
