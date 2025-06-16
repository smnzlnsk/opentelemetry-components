package oakestraheuristicengine

import (
	"errors"

	"github.com/smnzlnsk/opentelemetry-components/pkg/config"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

// Config defines the configuration for the oakestraheuristicengine processor.
type Config struct {
	// Add your configuration fields here
	HTTPServer     config.HTTPServerConfig     `mapstructure:"http_server"`
	MongoDB        config.MongoDBConfig        `mapstructure:"mongodb"`
	ServiceManager config.ServiceManagerConfig `mapstructure:"service_manager"`
}

var _ component.Config = (*Config)(nil)
var _ confmap.Unmarshaler = (*Config)(nil)

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	// Validate HTTP server configuration if enabled
	if cfg.HTTPServer.Enabled {
		if cfg.HTTPServer.Port <= 0 || cfg.HTTPServer.Port > 65535 {
			return errors.New("http_server.port must be between 1 and 65535")
		}

		if cfg.HTTPServer.Host == "" {
			// Default to all interfaces if not specified
			cfg.HTTPServer.Host = "0.0.0.0"
		}
	}

	// Validate MongoDB configuration
	if cfg.MongoDB.Host == "" {
		return errors.New("mongodb.host is required")
	}

	if cfg.MongoDB.Port <= 0 || cfg.MongoDB.Port > 65535 {
		return errors.New("mongodb.port must be between 1 and 65535")
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
	return nil
}
