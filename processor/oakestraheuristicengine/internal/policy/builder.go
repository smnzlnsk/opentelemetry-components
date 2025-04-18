package policy

import (
	"github.com/Knetic/govaluate"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
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
	notificationInterfaceBuilder domain.NotificationInterfaceBuilder
	capabilities                 []domain.NotificationInterfaceCapability
	notificationInterfaces       map[domain.NotificationInterfaceCapability]domain.NotificationInterface
	heuristicEntity              domain.HeuristicEntity
}

var _ domain.PolicyBuilder = &policyBuilder{}

func NewPolicyBuilder() domain.PolicyBuilder {
	return &policyBuilder{
		notificationInterfaceBuilder: notification_interface.NewNotificationInterfaceBuilder(),
		notificationInterfaces:       make(map[domain.NotificationInterfaceCapability]domain.NotificationInterface),
	}
}

func (b *policyBuilder) WithName(name string) domain.PolicyBuilder {
	if b.name != "" {
		return b
	}
	b.name = name
	return b
}

func (b *policyBuilder) WithHeuristicEngine(engine domain.HeuristicEntity) domain.PolicyBuilder {
	b.heuristicEntity = engine
	return b
}

func (b *policyBuilder) WithPreEvaluationCondition(conditions string) domain.PolicyBuilder {
	expression, err := govaluate.NewEvaluableExpression(conditions)
	if err != nil {
		return b
	}
	b.preEvaluationConditions = append(b.preEvaluationConditions, expression)
	return b
}

func (b *policyBuilder) WithEvaluationCondition(condition string) domain.PolicyBuilder {
	expression, err := govaluate.NewEvaluableExpression(condition)
	if err != nil {
		return b
	}
	b.evaluationConditions = append(b.evaluationConditions, expression)
	return b
}

func (b *policyBuilder) WithRoute(measure domain.NotificationInterface) domain.PolicyBuilder {
	if b.notificationInterfaces[domain.NotificationInterfaceCapability_Route] != nil {
		return b
	}
	if measure.Type() != domain.NotificationInterfaceCapability_Route {
		return b
	}
	b.capabilities = append(b.capabilities, domain.NotificationInterfaceCapability_Route)
	b.notificationInterfaces[domain.NotificationInterfaceCapability_Route] = measure
	return b
}

func (b *policyBuilder) WithRouteCondition(condition string) domain.PolicyBuilder {
	expression, err := govaluate.NewEvaluableExpression(condition)
	if err != nil {
		return b
	}
	b.routeConditions = append(b.routeConditions, expression)
	return b
}

func (b *policyBuilder) WithAlert(measure domain.NotificationInterface) domain.PolicyBuilder {
	if b.notificationInterfaces[domain.NotificationInterfaceCapability_Alert] != nil {
		return b
	}
	if measure.Type() != domain.NotificationInterfaceCapability_Alert {
		return b
	}
	b.capabilities = append(b.capabilities, domain.NotificationInterfaceCapability_Alert)
	b.notificationInterfaces[domain.NotificationInterfaceCapability_Alert] = measure
	return b
}

func (b *policyBuilder) WithAlertCondition(condition string) domain.PolicyBuilder {
	expression, err := govaluate.NewEvaluableExpression(condition)
	if err != nil {
		return b
	}
	b.alertConditions = append(b.alertConditions, expression)
	return b
}

func (b *policyBuilder) WithSchedule(measure domain.NotificationInterface) domain.PolicyBuilder {
	if b.notificationInterfaces[domain.NotificationInterfaceCapability_Schedule] != nil {
		return b
	}
	if measure.Type() != domain.NotificationInterfaceCapability_Schedule {
		return b
	}
	b.capabilities = append(b.capabilities, domain.NotificationInterfaceCapability_Schedule)
	b.notificationInterfaces[domain.NotificationInterfaceCapability_Schedule] = measure
	return b
}

func (b *policyBuilder) WithScheduleCondition(condition string) domain.PolicyBuilder {
	expression, err := govaluate.NewEvaluableExpression(condition)
	if err != nil {
		return b
	}
	b.scheduleConditions = append(b.scheduleConditions, expression)
	return b
}

func (b *policyBuilder) WithHeuristicEntity(entity domain.HeuristicEntity) domain.PolicyBuilder {
	b.heuristicEntity = entity
	return b
}

func (b *policyBuilder) Build() domain.Policy {
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
	if _, ok := b.notificationInterfaces[domain.NotificationInterfaceCapability_Alert]; ok {
		if len(b.alertConditions) == 0 {
			return nil
		}
	}

	if _, ok := b.notificationInterfaces[domain.NotificationInterfaceCapability_Route]; ok {
		if len(b.routeConditions) == 0 {
			return nil
		}
	}

	if _, ok := b.notificationInterfaces[domain.NotificationInterfaceCapability_Schedule]; ok {
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
	b.notificationInterfaces = make(map[domain.NotificationInterfaceCapability]domain.NotificationInterface)
	b.heuristicEntity = nil
	b.preEvaluationConditions = nil
	b.evaluationConditions = nil
	b.alertConditions = nil
	b.routeConditions = nil
	b.scheduleConditions = nil

	return policy
}

func (b *policyBuilder) NotificationInterfaceBuilder() domain.NotificationInterfaceBuilder {
	return b.notificationInterfaceBuilder
}
