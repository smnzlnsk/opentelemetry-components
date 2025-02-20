package policy

import (
	"fmt"

	"github.com/Knetic/govaluate"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/constants"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/notification_interface/factory"
)

// policyBuilder implements interfaces.PolicyBuilder
type policyBuilder struct {
	name                         string
	heuristicentity              interfaces.HeuristicEntity
	preEvaluationConditions      []*govaluate.EvaluableExpression
	evaluationConditions         []*govaluate.EvaluableExpression
	alertConditions              []*govaluate.EvaluableExpression
	routeConditions              []*govaluate.EvaluableExpression
	scheduleConditions           []*govaluate.EvaluableExpression
	notificationInterfaceFactory interfaces.NotificationInterfaceFactory
	capabilities                 []types.NotificationInterfaceCapability
	notificationInterfaces       map[types.NotificationInterfaceCapability]interfaces.NotificationInterface
	heuristicEntity              interfaces.HeuristicEntity
}

var _ interfaces.PolicyBuilder = &policyBuilder{}

func NewPolicyBuilder() interfaces.PolicyBuilder {
	return &policyBuilder{
		notificationInterfaceFactory: factory.NewNotificationInterfaceFactory(nil),
		notificationInterfaces:       make(map[types.NotificationInterfaceCapability]interfaces.NotificationInterface),
	}
}

func (b *policyBuilder) WithName(name string) interfaces.PolicyBuilder {
	if b.name != "" {
		return b
	}
	b.name = name
	return b
}

func (b *policyBuilder) WithHeuristicEngine(engine interfaces.HeuristicEntity) interfaces.PolicyBuilder {
	b.heuristicEntity = engine
	return b
}

func (b *policyBuilder) WithPreEvaluationCondition(conditions string) interfaces.PolicyBuilder {
	expression, err := govaluate.NewEvaluableExpression(conditions)
	if err != nil {
		return b
	}
	b.preEvaluationConditions = append(b.preEvaluationConditions, expression)
	return b
}

func (b *policyBuilder) WithEvaluationCondition(condition string) interfaces.PolicyBuilder {
	expression, err := govaluate.NewEvaluableExpression(condition)
	if err != nil {
		return b
	}
	b.evaluationConditions = append(b.evaluationConditions, expression)
	return b
}

func (b *policyBuilder) WithRoute(measure interfaces.NotificationInterface) interfaces.PolicyBuilder {
	if b.notificationInterfaces[constants.NotificationInterfaceCapability_Route] != nil {
		return b
	}
	b.capabilities = append(b.capabilities, constants.NotificationInterfaceCapability_Route)
	b.notificationInterfaces[constants.NotificationInterfaceCapability_Route] = measure
	return b
}

func (b *policyBuilder) WithRouteCondition(condition string) interfaces.PolicyBuilder {
	expression, err := govaluate.NewEvaluableExpression(condition)
	if err != nil {
		return b
	}
	b.routeConditions = append(b.routeConditions, expression)
	return b
}

func (b *policyBuilder) WithAlert(measure interfaces.NotificationInterface) interfaces.PolicyBuilder {
	if b.notificationInterfaces[constants.NotificationInterfaceCapability_Alert] != nil {
		return b
	}
	b.capabilities = append(b.capabilities, constants.NotificationInterfaceCapability_Alert)
	b.notificationInterfaces[constants.NotificationInterfaceCapability_Alert] = measure
	return b
}

func (b *policyBuilder) WithAlertCondition(condition string) interfaces.PolicyBuilder {
	expression, err := govaluate.NewEvaluableExpression(condition)
	if err != nil {
		return b
	}
	b.alertConditions = append(b.alertConditions, expression)
	return b
}

func (b *policyBuilder) WithSchedule(measure interfaces.NotificationInterface) interfaces.PolicyBuilder {
	if b.notificationInterfaces[constants.NotificationInterfaceCapability_Schedule] != nil {
		return b
	}
	b.capabilities = append(b.capabilities, constants.NotificationInterfaceCapability_Schedule)
	b.notificationInterfaces[constants.NotificationInterfaceCapability_Schedule] = measure
	return b
}

func (b *policyBuilder) WithScheduleCondition(condition string) interfaces.PolicyBuilder {
	expression, err := govaluate.NewEvaluableExpression(condition)
	if err != nil {
		return b
	}
	b.scheduleConditions = append(b.scheduleConditions, expression)
	return b
}

func (b *policyBuilder) WithHeuristicEntity(entity interfaces.HeuristicEntity) interfaces.PolicyBuilder {
	b.heuristicEntity = entity
	return b
}

func (b *policyBuilder) Build() interfaces.Policy {
	if b.name == "" {
		return nil
	}

	if len(b.capabilities) == 0 {
		fmt.Println("capabilities are not set")
		return nil
	}

	if len(b.notificationInterfaces) == 0 {
		fmt.Println("notification interfaces are not set")
		return nil
	}

	if b.heuristicEntity == nil {
		fmt.Println("heuristic entity is not set")
		return nil
	}

	if len(b.preEvaluationConditions) == 0 {
		fmt.Println("pre evaluation conditions are not set")
		return nil
	}

	if len(b.evaluationConditions) == 0 {
		fmt.Println("evaluation conditions are not set")
		return nil
	}

	// verify, if notification interfaces are set, then at least one condition is set
	if _, ok := b.notificationInterfaces[constants.NotificationInterfaceCapability_Alert]; ok {
		if len(b.alertConditions) == 0 {
			fmt.Println("alert conditions are not set")
			return nil
		}
	}

	if _, ok := b.notificationInterfaces[constants.NotificationInterfaceCapability_Route]; ok {
		if len(b.routeConditions) == 0 {
			fmt.Println("route conditions are not set")
			return nil
		}
	}

	if _, ok := b.notificationInterfaces[constants.NotificationInterfaceCapability_Schedule]; ok {
		if len(b.scheduleConditions) == 0 {
			fmt.Println("schedule conditions are not set")
			return nil
		}
	}

	policy := &policy{
		name:                    b.name,
		heuristicEntity:         b.heuristicentity,
		capabilities:            b.capabilities,
		notificationInterfaces:  b.notificationInterfaces,
		preEvaluationConditions: b.preEvaluationConditions,
		evaluationConditions:    b.evaluationConditions,
		alertConditions:         b.alertConditions,
		routeConditions:         b.routeConditions,
		scheduleConditions:      b.scheduleConditions,
	}

	// Reset internal state
	b.name = ""
	b.capabilities = nil
	b.notificationInterfaces = make(map[types.NotificationInterfaceCapability]interfaces.NotificationInterface)

	return policy
}

func (b *policyBuilder) NotificationInterfaceFactory() interfaces.NotificationInterfaceFactory {
	return b.notificationInterfaceFactory
}
