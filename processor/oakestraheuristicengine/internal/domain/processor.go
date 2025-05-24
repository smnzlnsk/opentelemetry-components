package domain

import (
	"github.com/Knetic/govaluate"
	"github.com/smnzlnsk/opentelemetry-components/internal/shared/evaluation"
)

type Processor interface {
	Identifier() string
	Process(instanceNumber int, prev float64, params map[string]interface{}) (evaluation.Entry, error)
	Evaluator() Evaluator
	GetNotificationCondition(capability NotificationInterfaceCapability) *govaluate.EvaluableExpression
}
