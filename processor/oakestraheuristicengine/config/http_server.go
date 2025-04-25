package config

import (
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

// HTTPServerConfig defines the configuration for the HTTP server
type HTTPServerConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Host    string `mapstructure:"host"`
}

var _ component.Config = (*HTTPServerConfig)(nil)
var _ confmap.Unmarshaler = (*HTTPServerConfig)(nil)

func (c *HTTPServerConfig) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	if c.Host == "" {
		return fmt.Errorf("host is required")
	}

	return nil
}

func (c *HTTPServerConfig) Unmarshal(cp *confmap.Conf) error {
	return cp.Unmarshal(c, confmap.WithIgnoreUnused())
}
