package policy

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/notification_interface"
)

// policyBuilder implements interfaces.PolicyBuilder
type policyBuilder struct {
	name                         string
	notificationInterfaceBuilder domain.NotificationInterfaceBuilder[any]
	capabilities                 []domain.NotificationInterfaceCapability
	notificationInterfaces       map[domain.NotificationInterfaceCapability]domain.NotificationInterface[any]
	heuristicEntity              domain.HeuristicEntity
}

var _ domain.PolicyBuilder = &policyBuilder{}

func NewPolicyBuilder() domain.PolicyBuilder {
	return &policyBuilder{
		notificationInterfaceBuilder: notification_interface.NewNotificationInterfaceBuilder[any](),
		notificationInterfaces:       make(map[domain.NotificationInterfaceCapability]domain.NotificationInterface[any]),
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

func (b *policyBuilder) WithRouteInterface(measure domain.NotificationInterface[any]) domain.PolicyBuilder {
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

func (b *policyBuilder) WithAlertInterface(measure domain.NotificationInterface[any]) domain.PolicyBuilder {
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

func (b *policyBuilder) WithScheduleInterface(measure domain.NotificationInterface[any]) domain.PolicyBuilder {
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

	// Set the notification interfaces on the heuristic entity
	for _, capability := range b.capabilities {
		switch capability {
		case domain.NotificationInterfaceCapability_Alert:
			b.heuristicEntity.SetAlert(b.notificationInterfaces[capability])
		case domain.NotificationInterfaceCapability_Route:
			b.heuristicEntity.SetRoute(b.notificationInterfaces[capability])
		case domain.NotificationInterfaceCapability_Schedule:
			b.heuristicEntity.SetSchedule(b.notificationInterfaces[capability])
		}
	}

	policy := &policy{
		name:            b.name,
		heuristicEntity: b.heuristicEntity,
	}

	// Reset internal state
	b.name = ""
	b.capabilities = nil
	b.notificationInterfaces = make(map[domain.NotificationInterfaceCapability]domain.NotificationInterface[any])
	b.heuristicEntity = nil

	return policy
}

func (b *policyBuilder) NotificationInterfaceBuilder() domain.NotificationInterfaceBuilder[any] {
	return b.notificationInterfaceBuilder
}
