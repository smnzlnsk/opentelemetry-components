package interfaces

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
)

type NotificationInterfaceFactory interface {
	CreateNotificationInterfaceBuilder(interfaceType types.NotificationInterfaceCapability) NotificationInterfaceBuilder
}

type NotificationInterfaceBuilder interface {
	WithHost(host string) NotificationInterfaceBuilder
	WithPort(port int) NotificationInterfaceBuilder
	WithEndpoint(endpoint string) NotificationInterfaceBuilder
	Build() NotificationInterface
}

type NotificationInterfaceRegistry interface {
	Register(notification NotificationInterface)
	Get(name types.NotificationInterfaceCapability) NotificationInterface
}

type NotificationInterface interface {
	Notify() error
	Type() types.NotificationInterfaceCapability
}
