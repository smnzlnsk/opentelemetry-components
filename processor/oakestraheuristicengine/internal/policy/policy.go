package policy

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
)

type Policy interface {
	Enforce() error
	Capabilities() []types.PolicyNotificationCapability
	Name() string
	GetMeasure(capability types.PolicyNotificationCapability) interfaces.MeasureNotifier
}

type policy struct {
	name            string
	capabilities    []types.PolicyNotificationCapability
	measures        map[types.PolicyNotificationCapability]interfaces.MeasureNotifier
	heuristicEntity interfaces.HeuristicEntity
}

func NewPolicy(
	name string,
	capabilities []types.PolicyNotificationCapability,
	measures map[types.PolicyNotificationCapability]interfaces.MeasureNotifier,
	heuristicEntity interfaces.HeuristicEntity,
) Policy {
	return &policy{
		name:            name,
		capabilities:    capabilities,
		measures:        measures,
		heuristicEntity: heuristicEntity,
	}
}

func (p *policy) Enforce() error {
	// TODO: implement enforcing of the policy
	return p.heuristicEntity.EvaluatePolicy()
}

func (p *policy) Capabilities() []types.PolicyNotificationCapability {
	return p.capabilities
}

func (p *policy) Name() string {
	return p.name
}

func (p *policy) GetMeasure(capability types.PolicyNotificationCapability) interfaces.MeasureNotifier {
	return p.measures[capability]
}
