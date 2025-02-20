package interfaces

import "github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"

type PolicyBuilder interface {
	WithName(name string) PolicyBuilder
	WithHeuristicEngine(engine HeuristicEntity) PolicyBuilder
	WithPreEvaluationCondition(condition string) PolicyBuilder
	WithEvaluationCondition(condition string) PolicyBuilder
	WithAlertCondition(condition string) PolicyBuilder
	WithRouteCondition(condition string) PolicyBuilder
	WithScheduleCondition(condition string) PolicyBuilder
	WithRoute(measure NotificationInterface) PolicyBuilder
	WithAlert(measure NotificationInterface) PolicyBuilder
	WithSchedule(measure NotificationInterface) PolicyBuilder
	WithHeuristicEntity(entity HeuristicEntity) PolicyBuilder
	NotificationInterfaceFactory() NotificationInterfaceFactory
	Build() Policy
}

type Policy interface {
	Enforce(values map[string]interface{}) error
	Capabilities() []types.NotificationInterfaceCapability
	Name() string
	GetNotificationInterface(capability types.NotificationInterfaceCapability) NotificationInterface
}
