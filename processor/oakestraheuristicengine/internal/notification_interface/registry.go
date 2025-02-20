package notification_interface

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
	"go.uber.org/zap"
)

type notificationInterfaceRegistry struct {
	logger                 *zap.Logger
	notificationInterfaces map[types.NotificationInterfaceCapability]interfaces.NotificationInterface
}

var _ interfaces.NotificationInterfaceRegistry = (*notificationInterfaceRegistry)(nil)

func NewNotificationInterfaceRegistry(logger *zap.Logger) interfaces.NotificationInterfaceRegistry {
	return &notificationInterfaceRegistry{
		logger:                 logger,
		notificationInterfaces: make(map[types.NotificationInterfaceCapability]interfaces.NotificationInterface),
	}
}

func (r *notificationInterfaceRegistry) Register(notificationInterface interfaces.NotificationInterface) {
	r.notificationInterfaces[notificationInterface.Type()] = notificationInterface
}

func (r *notificationInterfaceRegistry) Get(name types.NotificationInterfaceCapability) interfaces.NotificationInterface {
	return r.notificationInterfaces[name]
}
