package notification_interface

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"go.uber.org/zap"
)

type notificationInterfaceRegistry struct {
	logger                 *zap.Logger
	notificationInterfaces map[domain.NotificationInterfaceCapability]domain.NotificationInterface
}

var _ domain.NotificationInterfaceRegistry = (*notificationInterfaceRegistry)(nil)

func NewNotificationInterfaceRegistry(logger *zap.Logger) domain.NotificationInterfaceRegistry {
	return &notificationInterfaceRegistry{
		logger:                 logger,
		notificationInterfaces: make(map[domain.NotificationInterfaceCapability]domain.NotificationInterface),
	}
}

func (r *notificationInterfaceRegistry) Register(notificationInterface domain.NotificationInterface) {
	r.notificationInterfaces[notificationInterface.Type()] = notificationInterface
}

func (r *notificationInterfaceRegistry) Get(name domain.NotificationInterfaceCapability) domain.NotificationInterface {
	return r.notificationInterfaces[name]
}
