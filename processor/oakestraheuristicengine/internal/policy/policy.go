package policy

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/heuristicentity/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/measure"
)

type Policy interface {
	Enforce() map[string]measure.Measure
}

type policy struct {
	ID         string
	PolicyName string
	enforcer   interfaces.HeuristicEntity
}

func NewPolicy(id string, policyName string, enforcer interfaces.HeuristicEntity) Policy {
	return &policy{
		ID:         id,
		PolicyName: policyName,
		enforcer:   enforcer,
	}
}

func (p *policy) Enforce() map[string]measure.Measure {
	return p.enforcer.EvaluatePolicy()
}
