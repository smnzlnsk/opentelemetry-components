package notification_interface

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
)

type notificationInterfaceBuilder[T any] struct {
	notificationInterface *notificationInterface[T]
}

func NewNotificationInterfaceBuilder[T any]() domain.NotificationInterfaceBuilder[T] {
	return &notificationInterfaceBuilder[T]{
		notificationInterface: &notificationInterface[T]{},
	}
}

func (b *notificationInterfaceBuilder[T]) WithHost(host string) domain.NotificationInterfaceBuilder[T] {
	b.notificationInterface.host = host
	return b
}

func (b *notificationInterfaceBuilder[T]) WithPort(port int) domain.NotificationInterfaceBuilder[T] {
	b.notificationInterface.port = port
	return b
}

func (b *notificationInterfaceBuilder[T]) WithEndpoint(endpoint string) domain.NotificationInterfaceBuilder[T] {
	b.notificationInterface.endpoint = endpoint
	return b
}

func (b *notificationInterfaceBuilder[T]) WithCapability(capability domain.NotificationInterfaceCapability) domain.NotificationInterfaceBuilder[T] {
	b.notificationInterface.capability = capability
	return b
}

func (b *notificationInterfaceBuilder[T]) Build() domain.NotificationInterface[T] {
	ni := b.notificationInterface

	// reset builder state
	b.notificationInterface = &notificationInterface[T]{}

	return ni
}
