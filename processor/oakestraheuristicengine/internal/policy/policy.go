package policy

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
)

// policy implements interfaces.Policy
type policy struct {
	name            string
	heuristicEntity domain.HeuristicEntity
}

var _ domain.Policy = &policy{}

func (p *policy) Enforce(arguments ...interface{}) error {
	err := p.heuristicEntity.Evaluate(arguments...)
	if err != nil {
		return err
	}
	return nil
}

func (p *policy) Name() string {
	return p.name
}

func (p *policy) HeuristicEntity() domain.HeuristicEntity {
	return p.heuristicEntity
}
