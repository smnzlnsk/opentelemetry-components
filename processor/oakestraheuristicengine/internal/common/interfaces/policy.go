package interfaces

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
)

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
	NotificationInterfaceBuilder() NotificationInterfaceBuilder
	Build() Policy
}

type Policy interface {
	Check(values map[string]interface{}) error
	Enforce(values map[string]interface{}) error
	CheckPreEvaluationCondition(values map[string]interface{}) error
	CheckEvaluationCondition(values map[string]interface{}) error
	CheckNotificationConditions(values map[string]interface{}) error
	Capabilities() []types.NotificationInterfaceCapability
	Name() string
	HeuristicEngine() HeuristicEntity
	NotificationInterface(capability types.NotificationInterfaceCapability) NotificationInterface
}
