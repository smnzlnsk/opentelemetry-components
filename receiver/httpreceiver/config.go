package httpreceiver // import github.com/smnzlnsk/opentelemetry-components/receiver/httpreceiver
import (
	"errors"
	"net"
)

// Config represents the receiver config settings within the collector's config.yaml
type Config struct {
	endpoint endpoint `mapstructure:"endpoint"`
}

type endpoint struct {
	ip   string `mapstructure:"ip"`
	port int    `mapstructure:"port"`
}

// Validate checks if the receiver configuration is valid
// Additionally it overrides the configuration set if the respective environment variables are set
func (cfg *Config) Validate() error {
	if cfg.endpoint.ip == "" {
		return errors.New("endpoint.ip is required")
	}
	if net.ParseIP(cfg.endpoint.ip) == nil {
		return errors.New("endpoint.ip is not a valid IP address")
	}
	if cfg.endpoint.port > 65535 || cfg.endpoint.port < 1024 {
		return errors.New("endpoint.port is not a valid port")
	}
	return nil
}
