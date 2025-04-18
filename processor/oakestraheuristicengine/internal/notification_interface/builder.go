package notification_interface

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
)

type notificationInterfaceBuilder struct {
	capability domain.NotificationInterfaceCapability
	host       string
	port       int
	endpoint   string
}

func NewNotificationInterfaceBuilder() domain.NotificationInterfaceBuilder {
	return &notificationInterfaceBuilder{}
}

func (b *notificationInterfaceBuilder) WithHost(host string) domain.NotificationInterfaceBuilder {
	b.host = host
	return b
}

func (b *notificationInterfaceBuilder) WithPort(port int) domain.NotificationInterfaceBuilder {
	b.port = port
	return b
}

func (b *notificationInterfaceBuilder) WithEndpoint(endpoint string) domain.NotificationInterfaceBuilder {
	b.endpoint = endpoint
	return b
}

func (b *notificationInterfaceBuilder) WithCapability(capability domain.NotificationInterfaceCapability) domain.NotificationInterfaceBuilder {
	b.capability = capability
	return b
}

func (b *notificationInterfaceBuilder) Build() domain.NotificationInterface {
	ni := &notificationInterface{
		capability: b.capability,
		host:       b.host,
		port:       b.port,
		endpoint:   b.endpoint,
	}

	// reset builder state
	b.capability = domain.NotificationInterfaceCapability_Unknown
	b.host = ""
	b.port = 0
	b.endpoint = ""

	return ni
}
