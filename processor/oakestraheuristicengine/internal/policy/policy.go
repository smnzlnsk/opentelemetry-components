package policy

import (
	"github.com/Knetic/govaluate"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/constants"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
)

// policy implements interfaces.Policy
type policy struct {
	name                    string
	capabilities            []types.NotificationInterfaceCapability
	notificationInterfaces  map[types.NotificationInterfaceCapability]interfaces.NotificationInterface
	heuristicEntity         interfaces.HeuristicEntity
	preEvaluationConditions []*govaluate.EvaluableExpression
	evaluationConditions    []*govaluate.EvaluableExpression
	alertConditions         []*govaluate.EvaluableExpression
	routeConditions         []*govaluate.EvaluableExpression
	scheduleConditions      []*govaluate.EvaluableExpression
}

func (p *policy) CheckPreEvaluationCondition() error {
	for _, condition := range p.preEvaluationConditions {
		result, err := condition.Evaluate(nil)
		if err != nil {
			return err
		}
		if result != true {
			return nil
		}
	}
	return nil
}
func (p *policy) CheckEvaluationCondition() error {
	for _, condition := range p.evaluationConditions {
		result, err := condition.Evaluate(nil)
		if err != nil {
			return err
		}
		if result != true {
			return nil
		}
	}
	return nil
}

func (p *policy) CheckNotificationConditions(result map[string]interface{}) error {
	for _, condition := range p.alertConditions {
		result, err := condition.Evaluate(result)
		if err != nil {
			return err
		}
		if result == true {
			return p.GetNotificationInterface(constants.NotificationInterfaceCapability_Alert).Notify()
		}
	}

	for _, condition := range p.routeConditions {
		result, err := condition.Evaluate(result)
		if err != nil {
			return err
		}
		if result == true {
			return p.GetNotificationInterface(constants.NotificationInterfaceCapability_Route).Notify()
		}
	}

	for _, condition := range p.scheduleConditions {
		result, err := condition.Evaluate(result)
		if err != nil {
			return err
		}
		if result == true {
			return p.GetNotificationInterface(constants.NotificationInterfaceCapability_Schedule).Notify()
		}
	}
	return nil
}
func (p *policy) Enforce(values map[string]interface{}) error {
	result := p.heuristicEntity.Evaluate(values)
	return p.CheckNotificationConditions(result)
}

func (p *policy) Capabilities() []types.NotificationInterfaceCapability {
	return p.capabilities
}

func (p *policy) Name() string {
	return p.name
}

func (p *policy) GetNotificationInterface(capability types.NotificationInterfaceCapability) interfaces.NotificationInterface {
	return p.notificationInterfaces[capability]
}
