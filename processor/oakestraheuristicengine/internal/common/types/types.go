package types

type HeuristicType string

type NotificationInterfaceCapability int

func (c NotificationInterfaceCapability) String() string {
	return [...]string{
		"alert",
		"schedule",
		"route",
	}[c]
}
