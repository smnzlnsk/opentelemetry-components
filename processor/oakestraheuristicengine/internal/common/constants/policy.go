package constants

import "github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"

const (
	PolicyNotificationCapability_Unknown types.PolicyNotificationCapability = iota
	PolicyNotificationCapability_Alert
	PolicyNotificationCapability_Schedule
	PolicyNotificationCapability_Route
)
