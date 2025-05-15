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

type NotificationInterfaceBuilder[T any] interface {
	WithHost(host string) NotificationInterfaceBuilder[T]
	WithPort(port int) NotificationInterfaceBuilder[T]
	WithEndpoint(endpoint string) NotificationInterfaceBuilder[T]
	WithCapability(capability NotificationInterfaceCapability) NotificationInterfaceBuilder[T]
	Build() NotificationInterface[T]
}

type NotificationInterfaceRegistry interface {
	Register(notification NotificationInterface[any])
	Get(name NotificationInterfaceCapability) NotificationInterface[any]
}

type NotificationInterface[T any] interface {
	Notify(notification T) error
	Type() NotificationInterfaceCapability
}
