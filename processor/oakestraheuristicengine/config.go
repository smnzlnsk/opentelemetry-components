package oakestraheuristicengine

import (
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

const (
	notificationInterfaceKey = "interfaces"
	httpServerKey            = "http_server"
)

// HTTPServerConfig defines the configuration for the HTTP server
type HTTPServerConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Host    string `mapstructure:"host"`
}

type InterfacesConfig struct {
	Alert    InterfaceConfig `mapstructure:"alert"`
	Route    InterfaceConfig `mapstructure:"route"`
	Schedule InterfaceConfig `mapstructure:"schedule"`
}

type InterfaceConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

// Config defines the configuration for the oakestraheuristicengine processor.
type Config struct {
	// Add your configuration fields here
	HTTPServer             HTTPServerConfig `mapstructure:"http_server"`
	NotificationInterfaces InterfacesConfig `mapstructure:"interfaces"`
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
