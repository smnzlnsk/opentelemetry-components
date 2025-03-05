package policy

import (
	"fmt"

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

var _ interfaces.Policy = &policy{}

func (p *policy) Check(values map[string]interface{}) error {
	err := p.CheckPreEvaluationCondition(values)
	if err != nil {
		return err
	}

	err = p.CheckEvaluationCondition(values)
	if err != nil {
		return err
	}

	err = p.Enforce(values)
	if err != nil {
		return err
	}

	return nil
}

// returns nil (= success), if one pre evaluation condition is met
func (p *policy) CheckPreEvaluationCondition(values map[string]interface{}) error {
	for _, condition := range p.preEvaluationConditions {
		result, err := condition.Evaluate(values)
		if err != nil {
			return err
		}
		if result == true {
			return nil
		}
	}

	return fmt.Errorf("pre evaluation conditions not met")
}

// returns nil (= success), if one evaluation condition is met
func (p *policy) CheckEvaluationCondition(values map[string]interface{}) error {
	for _, condition := range p.evaluationConditions {
		result, err := condition.Evaluate(values)
		if err != nil {
			return err
		}
		if result == true {
			return nil
		}
	}
	return fmt.Errorf("evaluation conditions not met")
}

func (p *policy) CheckNotificationConditions(values map[string]interface{}) error {

	for _, condition := range p.alertConditions {
		result, err := condition.Evaluate(values)
		if err != nil {
			return err
		}
		if result == true {
			return p.NotificationInterface(constants.NotificationInterfaceCapability_Alert).Notify()
		}
	}

	for _, condition := range p.routeConditions {
		result, err := condition.Evaluate(values)
		if err != nil {
			return err
		}
		if result == true {
			return p.NotificationInterface(constants.NotificationInterfaceCapability_Route).Notify()
		}
	}

	for _, condition := range p.scheduleConditions {
		result, err := condition.Evaluate(values)
		if err != nil {
			return err
		}
		if result == true {
			return p.NotificationInterface(constants.NotificationInterfaceCapability_Schedule).Notify()
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

func (p *policy) HeuristicEngine() interfaces.HeuristicEntity {
	return p.heuristicEntity
}

func (p *policy) NotificationInterface(capability types.NotificationInterfaceCapability) interfaces.NotificationInterface {
	return p.notificationInterfaces[capability]
}
