package notification_interface

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

type NotificationInterfaceConfig interface {
	Validate() error
}

type notificationInterfacesConfig struct {
	notificationInterfaces map[string]NotificationInterfaceConfig `mapstructure:"-"`
}

var _ component.Config = (*notificationInterfacesConfig)(nil)
var _ confmap.Unmarshaler = (*notificationInterfacesConfig)(nil)

func (c *notificationInterfacesConfig) Validate() error {
	return nil
}

func (c *notificationInterfacesConfig) Unmarshal(cp *confmap.Conf) error {
	if cp == nil {
		return nil
	}

	err := cp.Unmarshal(c, confmap.WithIgnoreUnused())
	if err != nil {
		return err
	}

	c.notificationInterfaces = make(map[string]NotificationInterfaceConfig)

	for key, value := range cp.ToStringMap() {
		c.notificationInterfaces[key] = value.(NotificationInterfaceConfig)
	}

	return nil
}
