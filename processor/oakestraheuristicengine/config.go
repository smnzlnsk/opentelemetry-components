package oakestraheuristicengine

import (
	"errors"

	"github.com/smnzlnsk/opentelemetry-components/pkg/config"
	"github.com/smnzlnsk/opentelemetry-components/pkg/database"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

// Config defines the configuration for the oakestraheuristicengine processor.
type Config struct {
	// Add your configuration fields here
	HTTPServer config.HTTPServerConfig `mapstructure:"http_server"`
	//NotificationInterfaces config.InterfacesConfig     `mapstructure:"interfaces"`
	Database       database.DatabaseConfig     `mapstructure:"database"`
	ServiceManager config.ServiceManagerConfig `mapstructure:"service_manager"`
}

var _ component.Config = (*Config)(nil)
var _ confmap.Unmarshaler = (*Config)(nil)

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	/*if len(cfg.NotificationInterfaces) == 0 {
		return errors.New("must provide at least one notification interface")
	}*/

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
	return nil
}
