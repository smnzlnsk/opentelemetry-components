package domain

import (
	"github.com/Knetic/govaluate"
	"github.com/smnzlnsk/opentelemetry-components/pkg/evaluation"
)

type Processor interface {
	Identifier() string
	Process(instanceNumber int, prev float64, params map[string]interface{}) (evaluation.Entry, error)
	Evaluator() Evaluator
	GetCapabilities() map[NotificationInterfaceCapability]bool
	GetNotificationCondition(capability NotificationInterfaceCapability) *govaluate.EvaluableExpression
}
