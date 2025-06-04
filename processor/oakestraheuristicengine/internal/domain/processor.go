package domain

import (
	"github.com/smnzlnsk/opentelemetry-components/pkg/evaluation"
	"github.com/smnzlnsk/opentelemetry-components/pkg/notification"
)

type Processor interface {
	Identifier() string
	Process(instanceNumber int, prev float64, params map[string]interface{}) (evaluation.Entry, error)
	Evaluator() Evaluator
	History() map[string]evaluation.Result
	GetCapabilities() map[NotificationInterfaceCapability]bool
	GetNotificationFunction(capability NotificationInterfaceCapability) notification.Function
}
