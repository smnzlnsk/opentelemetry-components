package config

import (
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

type InterfacesConfig struct {
	Alert    InterfaceConfig `mapstructure:"alert"`
	Route    InterfaceConfig `mapstructure:"route"`
	Schedule InterfaceConfig `mapstructure:"schedule"`
}

type InterfaceConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

var _ component.Config = (*InterfacesConfig)(nil)
var _ confmap.Unmarshaler = (*InterfacesConfig)(nil)

func (c *InterfacesConfig) Validate() error {
	if c.Alert.Port <= 0 || c.Alert.Port > 65535 {
		return fmt.Errorf("alert.port must be between 1 and 65535")
	}

	if c.Route.Port <= 0 || c.Route.Port > 65535 {
		return fmt.Errorf("route.port must be between 1 and 65535")
	}

	return nil
}

func (c *InterfacesConfig) Unmarshal(cp *confmap.Conf) error {
	return cp.Unmarshal(c, confmap.WithIgnoreUnused())
}
