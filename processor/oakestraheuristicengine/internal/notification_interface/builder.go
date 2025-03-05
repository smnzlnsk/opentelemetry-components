package notification_interface

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
)

type notificationInterfaceBuilder struct {
	capability types.NotificationInterfaceCapability
	host       string
	port       int
	endpoint   string
}

func NewNotificationInterfaceBuilder() interfaces.NotificationInterfaceBuilder {
	return &notificationInterfaceBuilder{}
}

func (b *notificationInterfaceBuilder) WithHost(host string) interfaces.NotificationInterfaceBuilder {
	b.host = host
	return b
}

func (b *notificationInterfaceBuilder) WithPort(port int) interfaces.NotificationInterfaceBuilder {
	b.port = port
	return b
}

func (b *notificationInterfaceBuilder) WithEndpoint(endpoint string) interfaces.NotificationInterfaceBuilder {
	b.endpoint = endpoint
	return b
}

func (b *notificationInterfaceBuilder) WithCapability(capability types.NotificationInterfaceCapability) interfaces.NotificationInterfaceBuilder {
	b.capability = capability
	return b
}

func (b *notificationInterfaceBuilder) Build() interfaces.NotificationInterface {
	ni := &notificationInterface{
		capability: b.capability,
		host:       b.host,
		port:       b.port,
		endpoint:   b.endpoint,
	}

	// reset builder state
	b.capability = types.NotificationInterfaceCapability_Unknown
	b.host = ""
	b.port = 0
	b.endpoint = ""

	return ni
}
