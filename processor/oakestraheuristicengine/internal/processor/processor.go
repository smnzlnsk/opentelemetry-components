package processor

import (
	"github.com/Knetic/govaluate"
	"github.com/smnzlnsk/opentelemetry-components/pkg/evaluation"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
)

// processor implements interfaces.Processor
type processor struct {
	identifier    string
	evaluator     domain.Evaluator
	resultHistory map[string]evaluation.Result

	// Notification conditions are set on a per processor basis
	notificationConditions map[domain.NotificationInterfaceCapability]*govaluate.EvaluableExpression
}

func NewProcessor(
	identifier string,
	evaluator domain.Evaluator,
	notificationConditions map[domain.NotificationInterfaceCapability]*govaluate.EvaluableExpression,
) domain.Processor {
	return &processor{
		identifier: identifier,
		evaluator:  evaluator,
	}
}

func (p *processor) Identifier() string {
	return p.identifier
}

func (p *processor) Evaluator() domain.Evaluator {
	return p.evaluator
}

func (p *processor) Process(instanceNumber int, prev float64, params map[string]interface{}) (evaluation.Entry, error) {
	priority, err := p.evaluator.Evaluate(prev, params)
	if err != nil {
		return evaluation.Entry{}, err
	}

	return evaluation.Entry{
		InstanceNumber: instanceNumber,
		Priority:       priority,
	}, nil
}

func (p *processor) GetNotificationCondition(capability domain.NotificationInterfaceCapability) *govaluate.EvaluableExpression {
	return p.notificationConditions[capability]
}

func (p *processor) GetCapabilities() map[domain.NotificationInterfaceCapability]bool {
	capabilities := make(map[domain.NotificationInterfaceCapability]bool)
	for capability, condition := range p.notificationConditions {
		capabilities[capability] = condition != nil
	}
	return capabilities
}
