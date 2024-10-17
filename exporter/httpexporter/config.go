package httpexporter // import github.com/smnzlnsk/opentelemetry-components/exporter/httpexporter
import (
	"errors"
	"net"
)

// Config represents the exporter config settings within the collector's config.yaml
type Config struct {
	Endpoint EndpointConfig `mapstructure:"endpoint"`
}

type EndpointConfig struct {
	IP   string `mapstructure:"ip"`
	Port int    `mapstructure:"port"`
}

// Validate checks if the receiver configuration is valid
// Additionally it overrides the configuration set if the respective environment variables are set
func (cfg *Config) Validate() error {
	if cfg.Endpoint.IP == "" {
		return errors.New("endpoint.ip is required")
	}
	if net.ParseIP(cfg.Endpoint.IP) == nil {
		return errors.New("endpoint.ip is not a valid IP address")
	}
	if cfg.Endpoint.Port > 65535 || cfg.Endpoint.Port < 1024 {
		return errors.New("endpoint.port is not a valid port")
	}
	return nil
}
