package domain

import "github.com/smnzlnsk/opentelemetry-components/pkg/notification"

type Evaluator interface {
	Evaluate(initalValue float64, params map[string]interface{}) (float64, error)
	EvaluationNotifier
}

type EvaluationNotifier interface {
	AlarmCondition() notification.Function
	RouteCondition() notification.Function
	ScheduleCondition() notification.Function
}
