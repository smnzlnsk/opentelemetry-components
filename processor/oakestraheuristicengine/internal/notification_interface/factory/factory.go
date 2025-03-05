package factory

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/notification_interface/implementations/alert"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/notification_interface/implementations/route"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/notification_interface/implementations/schedule"
	"go.uber.org/zap"
)

type notificationInterfaceFactory struct {
	logger     *zap.Logger
	interfaces map[types.NotificationInterfaceCapability]interfaces.NotificationInterface
}

func NewNotificationInterfaceFactory(logger *zap.Logger) interfaces.NotificationInterfaceFactory {
	return &notificationInterfaceFactory{
		logger:     logger,
		interfaces: make(map[types.NotificationInterfaceCapability]interfaces.NotificationInterface),
	}
}

func (f *notificationInterfaceFactory) CreateNotificationInterfaceBuilder(interfaceType types.NotificationInterfaceCapability) interfaces.NotificationInterfaceBuilder {
	switch interfaceType {
	case types.NotificationInterfaceCapability_Route:
		return route.NewRouteNotifierBuilder()
	case types.NotificationInterfaceCapability_Alert:
		return alert.NewAlertNotifierBuilder()
	case types.NotificationInterfaceCapability_Schedule:
		return schedule.NewScheduleNotifierBuilder()
	default:
		return nil
	}
}
