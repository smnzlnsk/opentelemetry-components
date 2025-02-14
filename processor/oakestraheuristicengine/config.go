package oakestraheuristicengine

import (
	"errors"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/measure"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

const (
	measureKey = "measures"
)

// Config defines the configuration for the oakestraheuristicengine processor.
type Config struct {
	// Add your configuration fields here
	Measures map[string]measure.MeasureConfig `mapstructure:"-"`
}

var _ component.Config = (*Config)(nil)
var _ confmap.Unmarshaler = (*Config)(nil)

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	if len(cfg.Measures) == 0 {
		return errors.New("must provide at least one measure")
	}
	return nil
}

func (cfg *Config) Unmarshal(cp *confmap.Conf) error {
	if cp == nil {
		return nil
	}

	err := cp.Unmarshal(cfg, confmap.WithIgnoreUnused())
	if err != nil {
		return err
	}

	cfg.Measures = map[string]measure.MeasureConfig{}

	measures, err := cp.Sub(measureKey)
	if err != nil {
		return err
	}

	for key, value := range measures.ToStringMap() {
		cfg.Measures[key] = value.(measure.MeasureConfig)
	}

	return nil
}
