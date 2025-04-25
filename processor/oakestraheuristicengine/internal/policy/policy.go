package policy

import (
	"fmt"

	"github.com/Knetic/govaluate"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
)

// policy implements interfaces.Policy
type policy struct {
	name                    string
	capabilities            []domain.NotificationInterfaceCapability
	notificationInterfaces  map[domain.NotificationInterfaceCapability]domain.NotificationInterface
	heuristicEntity         domain.HeuristicEntity
	preEvaluationConditions []*govaluate.EvaluableExpression
	evaluationConditions    []*govaluate.EvaluableExpression
	alertConditions         []*govaluate.EvaluableExpression
	routeConditions         []*govaluate.EvaluableExpression
	scheduleConditions      []*govaluate.EvaluableExpression
}

var _ domain.Policy = &policy{}

// Check checks if the policy is to be evaluated
// Returns nil if the policy is to be evaluated, otherwise an error is returned
func (p *policy) Check(values map[string]interface{}) error {
	err := p.CheckPreEvaluationCondition(values)
	if err != nil {
		return err
	}

	err = p.CheckEvaluationCondition(values)
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

func (p *policy) CheckNotificationConditions(evaluationResult domain.EvaluationResult) error {
	for _, result := range evaluationResult.Results {
		instanceName := fmt.Sprintf("%s.instance.%d", evaluationResult.JobName, result.InstanceNumber)
		fmt.Println(instanceName)
		values := evaluationResult.Values[instanceName]
		fmt.Println(evaluationResult.Values[instanceName])
		values["result"] = result.Priority
		for _, condition := range p.alertConditions {
			result, err := condition.Evaluate(values)
			if err != nil {
				return err
			}
			if result == true {
				return p.NotificationInterface(domain.NotificationInterfaceCapability_Alert).Notify(evaluationResult)
			}
		}

		for _, condition := range p.routeConditions {
			result, err := condition.Evaluate(values)
			if err != nil {
				return err
			}
			if result == true {
				return p.NotificationInterface(domain.NotificationInterfaceCapability_Route).Notify(evaluationResult)
			}
		}

		for _, condition := range p.scheduleConditions {
			result, err := condition.Evaluate(values)
			if err != nil {
				return err
			}
			if result == true {
				return p.NotificationInterface(domain.NotificationInterfaceCapability_Schedule).Notify(evaluationResult)
			}
		}
	}
	return nil
}
func (p *policy) Enforce(processorIdentifier string, arguments ...interface{}) error {
	evaluationResult := p.heuristicEntity.Evaluate(processorIdentifier, arguments...)
	return p.CheckNotificationConditions(evaluationResult)
}

func (p *policy) Capabilities() []domain.NotificationInterfaceCapability {
	return p.capabilities
}

func (p *policy) Name() string {
	return p.name
}

func (p *policy) HeuristicEngine() domain.HeuristicEntity {
	return p.heuristicEntity
}

func (p *policy) NotificationInterface(capability domain.NotificationInterfaceCapability) domain.NotificationInterface {
	return p.notificationInterfaces[capability]
}
