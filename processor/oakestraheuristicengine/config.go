package oakestraheuristicengine

import (
	"errors"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/notification_interface"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

const (
	notificationInterfaceKey = "interfaces"
)

// Config defines the configuration for the oakestraheuristicengine processor.
type Config struct {
	// Add your configuration fields here
	NotificationInterfaces map[string]notification_interface.NotificationInterfaceConfig `mapstructure:"-"`
}

var _ component.Config = (*Config)(nil)
var _ confmap.Unmarshaler = (*Config)(nil)

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	if len(cfg.NotificationInterfaces) == 0 {
		return errors.New("must provide at least one notification interface")
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

	cfg.NotificationInterfaces = map[string]notification_interface.NotificationInterfaceConfig{}

	notificationInterfaces, err := cp.Sub(notificationInterfaceKey)
	if err != nil {
		return err
	}

	for key, value := range notificationInterfaces.ToStringMap() {
		cfg.NotificationInterfaces[key] = value.(notification_interface.NotificationInterfaceConfig)
	}

	return nil
}
