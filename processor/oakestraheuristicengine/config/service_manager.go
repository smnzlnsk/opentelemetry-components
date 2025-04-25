package config

import (
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

type ServiceManagerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

var _ component.Config = (*ServiceManagerConfig)(nil)
var _ confmap.Unmarshaler = (*ServiceManagerConfig)(nil)

func (c *ServiceManagerConfig) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	if c.Host == "" {
		return fmt.Errorf("host is required")
	}

	return nil
}

func (c *ServiceManagerConfig) Unmarshal(cp *confmap.Conf) error {
	return cp.Unmarshal(c, confmap.WithIgnoreUnused())
}
