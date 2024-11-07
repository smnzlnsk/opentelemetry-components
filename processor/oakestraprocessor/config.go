package oakestraprocessor // import github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor

import (
	"errors"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
	"go.opentelemetry.io/collector/component"
)

var _ component.ConfigValidator = (*Config)(nil)

// Config represents the processor config settings within the collector's config.yaml
type Config struct {
	Processors map[string]internal.Config `mapstructure:"subprocessors"`
}

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	if len(cfg.Processors) == 0 {
		return errors.New("must provide at least one subprocessor")
	}
	return nil
}
