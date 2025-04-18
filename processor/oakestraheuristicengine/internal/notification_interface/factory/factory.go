package factory

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/notification_interface/implementations/alert"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/notification_interface/implementations/route"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/notification_interface/implementations/schedule"
	"go.uber.org/zap"
)

type notificationInterfaceFactory struct {
	logger     *zap.Logger
	interfaces map[domain.NotificationInterfaceCapability]domain.NotificationInterface
}

func NewNotificationInterfaceFactory(logger *zap.Logger) domain.NotificationInterfaceFactory {
	return &notificationInterfaceFactory{
		logger:     logger,
		interfaces: make(map[domain.NotificationInterfaceCapability]domain.NotificationInterface),
	}
}

func (f *notificationInterfaceFactory) CreateNotificationInterfaceBuilder(interfaceType domain.NotificationInterfaceCapability) domain.NotificationInterfaceBuilder {
	switch interfaceType {
	case domain.NotificationInterfaceCapability_Route:
		return route.NewRouteNotifierBuilder()
	case domain.NotificationInterfaceCapability_Alert:
		return alert.NewAlertNotifierBuilder()
	case domain.NotificationInterfaceCapability_Schedule:
		return schedule.NewScheduleNotifierBuilder()
	default:
		return nil
	}
}
