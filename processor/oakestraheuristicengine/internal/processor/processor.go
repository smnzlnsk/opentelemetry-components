package processor

import (
	"fmt"

	"github.com/smnzlnsk/opentelemetry-components/pkg/evaluation"
	"github.com/smnzlnsk/opentelemetry-components/pkg/notification"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
)

// processor implements interfaces.Processor
type processor struct {
	identifier    string
	evaluator     domain.Evaluator
	resultHistory map[string]evaluation.Result

	// Notification conditions are set on a per processor basis - they are imported on creation from the associated evaluator
	notificationConditions map[domain.NotificationInterfaceCapability]notification.Function
}

func NewProcessor(
	identifier string,
	evaluator domain.Evaluator,
) domain.Processor {
	p := &processor{
		identifier:             identifier,
		evaluator:              evaluator,
		notificationConditions: make(map[domain.NotificationInterfaceCapability]notification.Function),
		resultHistory:          make(map[string]evaluation.Result),
	}

	// Set the notification conditions on the processor
	// If the evaluator does not have a condition for a capability, it SHOULD return a nil function
	p.notificationConditions[domain.NotificationInterfaceCapability_Alert] = p.evaluator.AlarmCondition()
	p.notificationConditions[domain.NotificationInterfaceCapability_Route] = p.evaluator.RouteCondition()
	p.notificationConditions[domain.NotificationInterfaceCapability_Schedule] = p.evaluator.ScheduleCondition()

	return p
}

func (p *processor) Identifier() string {
	return p.identifier
}

func (p *processor) Evaluator() domain.Evaluator {
	return p.evaluator
}

func (p *processor) History() map[string]evaluation.Result {
	return p.resultHistory
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

func (p *processor) GetNotificationFunction(capability domain.NotificationInterfaceCapability) notification.Function {
	return p.notificationConditions[capability]
}

func (p *processor) GetCapabilities() map[domain.NotificationInterfaceCapability]bool {
	capabilities := make(map[domain.NotificationInterfaceCapability]bool)
	for capability, condition := range p.notificationConditions {
		capabilities[capability] = condition != nil
		if condition != nil {
			fmt.Println("capability", capability, "condition", condition)
		}
	}
	return capabilities
}
