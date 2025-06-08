package oakestraprocessor // import github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor

import (
	"errors"
	"fmt"

	"github.com/smnzlnsk/opentelemetry-components/pkg/database"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

const (
	processorKey = "subprocessors"
	mongodbKey   = "mongodb"
)

var (
	_ component.ConfigValidator = (*Config)(nil)
	_ component.Config          = (*Config)(nil)
	_ confmap.Unmarshaler       = (*Config)(nil)
)

// Config represents the processor config settings within the collector's config.yaml
type Config struct {
	Processors map[string]internal.Config `mapstructure:"-"`
	GRPCPort   int                        `mapstructure:"grpc_port"`
	Database   database.DatabaseConfig    `mapstructure:"database"`
}

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	if len(cfg.Processors) == 0 {
		return errors.New("must provide at least one subprocessor")
	}

	// Validate database configuration
	if cfg.Database.MongoDB == nil && cfg.Database.Redis == nil {
		return errors.New("database configuration is required - either 'mongodb' or 'redis' must be specified")
	}

	if cfg.Database.MongoDB != nil && cfg.Database.Redis != nil {
		return errors.New("only one database configuration can be specified - either 'mongodb' or 'redis', not both")
	}

	if cfg.Database.MongoDB != nil {
		if cfg.Database.MongoDB.Host == "" {
			return errors.New("database.mongodb.host is required")
		}
		if cfg.Database.MongoDB.Port <= 0 || cfg.Database.MongoDB.Port > 65535 {
			return errors.New("database.mongodb.port must be between 1 and 65535")
		}
	}

	if cfg.Database.Redis != nil {
		if cfg.Database.Redis.Host == "" {
			return errors.New("database.redis.host is required")
		}
		if cfg.Database.Redis.Port <= 0 || cfg.Database.Redis.Port > 65535 {
			return errors.New("database.redis.port must be between 1 and 65535")
		}
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
			return fmt.Errorf("error reading settings for processor %s: %w", key, err)
		}

		cfg.Processors[key] = processorCfg
	}
	return nil
}
