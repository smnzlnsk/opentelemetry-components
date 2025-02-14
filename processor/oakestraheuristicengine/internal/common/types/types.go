package types

type HeuristicType string
type MeasureType string

type PolicyNotificationCapability int

func (c PolicyNotificationCapability) String() string {
	return [...]string{
		"unknown",
		"alert",
		"schedule",
		"route",
	}[c]
}
