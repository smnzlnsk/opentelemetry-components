package config

import (
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

type MongoDBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

var _ component.Config = (*MongoDBConfig)(nil)
var _ confmap.Unmarshaler = (*MongoDBConfig)(nil)

func (c *MongoDBConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}

	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	return nil
}

func (c *MongoDBConfig) Unmarshal(cp *confmap.Conf) error {
	return cp.Unmarshal(c, confmap.WithIgnoreUnused())
}
