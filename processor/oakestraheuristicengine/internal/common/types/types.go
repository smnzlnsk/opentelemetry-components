package types

type HeuristicType string

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

type MetricValueType int

func (t MetricValueType) String() string {
	return [...]string{
		"raw",
		"sum",
		"avg",
		"count",
		"min",
		"max",
		"stddev",
		"median",
		"p90",
		"p95",
		"p99",
		"p999",
		"p9999",
	}[t]
}

type MetricKey struct {
	Name  string
	State string
	Type  MetricValueType
}
