package domain

type NotificationInterfaceCapability int

const (
	NotificationInterfaceCapability_Alert NotificationInterfaceCapability = iota
	NotificationInterfaceCapability_Schedule
	NotificationInterfaceCapability_Route
	NotificationInterfaceCapability_Unknown
)

func (c NotificationInterfaceCapability) String() string {
	return [...]string{
		"alert",
		"schedule",
		"route",
		"unknown",
	}[c]
}

type NotificationInterfaceFactory interface {
	CreateNotificationInterfaceBuilder(interfaceType NotificationInterfaceCapability) NotificationInterfaceBuilder
}

type NotificationInterfaceBuilder interface {
	WithHost(host string) NotificationInterfaceBuilder
	WithPort(port int) NotificationInterfaceBuilder
	WithEndpoint(endpoint string) NotificationInterfaceBuilder
	WithCapability(capability NotificationInterfaceCapability) NotificationInterfaceBuilder
	Build() NotificationInterface
}

type NotificationInterfaceRegistry interface {
	Register(notification NotificationInterface)
	Get(name NotificationInterfaceCapability) NotificationInterface
}

type NotificationInterface interface {
	Notify() error
	Type() NotificationInterfaceCapability
}
