package oakestraprocessor // import github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor

import (
	"errors"
	"fmt"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

const (
	processorKey = "subprocessors"
)

var (
	_ component.ConfigValidator = (*Config)(nil)
	_ component.Config          = (*Config)(nil)
	_ confmap.Unmarshaler       = (*Config)(nil)
)

// Config represents the processor config settings within the collector's config.yaml
type Config struct {
	Processors map[string]internal.Config `mapstructure:"-"`
}

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	if len(cfg.Processors) == 0 {
		return errors.New("must provide at least one subprocessor")
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

	cfg.Processors = map[string]internal.Config{}

	ps, err := cp.Sub(processorKey)
	if err != nil {
		return err
	}
	for key := range ps.ToStringMap() {
		factory, ok := getProcessorFactory(key)
		if !ok {
			return fmt.Errorf("invalid processor key: %s", key)
		}

		processorCfg := factory.CreateDefaultConfig()
		processorCfgSection, err := ps.Sub(key)
		if err != nil {
			return err
		}
		err = processorCfgSection.Unmarshal(processorCfg)
		if err != nil {
			return fmt.Errorf("Error reading settings for processor %s: %w", key, err)
		}

		cfg.Processors[key] = processorCfg
	}
	return nil
}
