package policy

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/constants"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/measure/factory"
)

type PolicyBuilder interface {
	WithName(name string) PolicyBuilder
	WithRoute(measure interfaces.MeasureNotifier) PolicyBuilder
	WithAlert(measure interfaces.MeasureNotifier) PolicyBuilder
	WithSchedule(measure interfaces.MeasureNotifier) PolicyBuilder
	WithHeuristicEntity(entity interfaces.HeuristicEntity) PolicyBuilder
	MeasureFactory() interfaces.MeasureFactory
	Build() Policy
}

type policyBuilder struct {
	name            string
	measureFactory  interfaces.MeasureFactory
	capabilities    []types.PolicyNotificationCapability
	measures        map[types.PolicyNotificationCapability]interfaces.MeasureNotifier
	heuristicEntity interfaces.HeuristicEntity
}

var _ PolicyBuilder = &policyBuilder{}

func NewPolicyBuilder() PolicyBuilder {
	return &policyBuilder{
		measureFactory: factory.NewMeasureFactory(nil),
		measures:       make(map[types.PolicyNotificationCapability]interfaces.MeasureNotifier),
	}
}

func (b *policyBuilder) WithName(name string) PolicyBuilder {
	if b.name != "" {
		return b
	}
	b.name = name
	return b
}

func (b *policyBuilder) WithRoute(measure interfaces.MeasureNotifier) PolicyBuilder {
	if b.measures[constants.PolicyNotificationCapability_Route] != nil {
		return b
	}
	b.capabilities = append(b.capabilities, constants.PolicyNotificationCapability_Route)
	b.measures[constants.PolicyNotificationCapability_Route] = measure
	return b
}

func (b *policyBuilder) WithAlert(measure interfaces.MeasureNotifier) PolicyBuilder {
	if b.measures[constants.PolicyNotificationCapability_Alert] != nil {
		return b
	}
	b.capabilities = append(b.capabilities, constants.PolicyNotificationCapability_Alert)
	b.measures[constants.PolicyNotificationCapability_Alert] = measure
	return b
}

func (b *policyBuilder) WithSchedule(measure interfaces.MeasureNotifier) PolicyBuilder {
	if b.measures[constants.PolicyNotificationCapability_Schedule] != nil {
		return b
	}
	b.capabilities = append(b.capabilities, constants.PolicyNotificationCapability_Schedule)
	b.measures[constants.PolicyNotificationCapability_Schedule] = measure
	return b
}

func (b *policyBuilder) WithHeuristicEntity(entity interfaces.HeuristicEntity) PolicyBuilder {
	b.heuristicEntity = entity
	return b
}

func (b *policyBuilder) Build() Policy {
	if b.name == "" {
		return nil
	}

	if len(b.capabilities) == 0 {
		return nil
	}

	if len(b.measures) == 0 {
		return nil
	}

	/* TODO: add later
	if b.heuristicEntity == nil {
		return nil
	}
	*/

	policy := &policy{
		name:            b.name,
		capabilities:    b.capabilities,
		measures:        b.measures,
		heuristicEntity: b.heuristicEntity,
	}

	// Reset internal state
	b.name = ""
	b.capabilities = nil
	b.measures = make(map[types.PolicyNotificationCapability]interfaces.MeasureNotifier)

	return policy
}

func (b *policyBuilder) MeasureFactory() interfaces.MeasureFactory {
	return b.measureFactory
}
