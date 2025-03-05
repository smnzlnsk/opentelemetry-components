package policy

import (
	"github.com/Knetic/govaluate"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/constants"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/notification_interface"
)

// policyBuilder implements interfaces.PolicyBuilder
type policyBuilder struct {
	name                         string
	preEvaluationConditions      []*govaluate.EvaluableExpression
	evaluationConditions         []*govaluate.EvaluableExpression
	alertConditions              []*govaluate.EvaluableExpression
	routeConditions              []*govaluate.EvaluableExpression
	scheduleConditions           []*govaluate.EvaluableExpression
	notificationInterfaceBuilder interfaces.NotificationInterfaceBuilder
	capabilities                 []types.NotificationInterfaceCapability
	notificationInterfaces       map[types.NotificationInterfaceCapability]interfaces.NotificationInterface
	heuristicEntity              interfaces.HeuristicEntity
}

var _ interfaces.PolicyBuilder = &policyBuilder{}

func NewPolicyBuilder() interfaces.PolicyBuilder {
	return &policyBuilder{
		notificationInterfaceBuilder: notification_interface.NewNotificationInterfaceBuilder(),
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
	if measure.Type() != constants.NotificationInterfaceCapability_Route {
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
	if measure.Type() != constants.NotificationInterfaceCapability_Alert {
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
	if measure.Type() != constants.NotificationInterfaceCapability_Schedule {
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
		return nil
	}

	if len(b.notificationInterfaces) == 0 {
		return nil
	}

	if b.heuristicEntity == nil {
		return nil
	}

	if len(b.preEvaluationConditions) == 0 {
		return nil
	}

	if len(b.evaluationConditions) == 0 {
		return nil
	}

	// verify, if notification interfaces are set, then at least one condition is set
	if _, ok := b.notificationInterfaces[constants.NotificationInterfaceCapability_Alert]; ok {
		if len(b.alertConditions) == 0 {
			return nil
		}
	}

	if _, ok := b.notificationInterfaces[constants.NotificationInterfaceCapability_Route]; ok {
		if len(b.routeConditions) == 0 {
			return nil
		}
	}

	if _, ok := b.notificationInterfaces[constants.NotificationInterfaceCapability_Schedule]; ok {
		if len(b.scheduleConditions) == 0 {
			return nil
		}
	}

	policy := &policy{
		name:                    b.name,
		heuristicEntity:         b.heuristicEntity,
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
	b.heuristicEntity = nil
	b.preEvaluationConditions = nil
	b.evaluationConditions = nil
	b.alertConditions = nil
	b.routeConditions = nil
	b.scheduleConditions = nil

	return policy
}

func (b *policyBuilder) NotificationInterfaceBuilder() interfaces.NotificationInterfaceBuilder {
	return b.notificationInterfaceBuilder
}
