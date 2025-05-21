package processor

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
)

// processor implements interfaces.Processor
type processor struct {
	identifier string
	evaluator  domain.Evaluator
}

func NewProcessor(identifier string, evaluator domain.Evaluator) domain.Processor {
	return &processor{
		identifier: identifier,
		evaluator:  evaluator,
	}
}

func (p *processor) Identifier() string {
	return p.identifier
}

func (p *processor) Evaluator() domain.Evaluator {
	return p.evaluator
}

func (p *processor) Process(instanceNumber int, prev float64, params map[string]interface{}) (domain.EvaluationEntry, error) {
	priority, err := p.evaluator.Evaluate(prev, params)
	if err != nil {
		return domain.EvaluationEntry{}, err
	}

	return domain.EvaluationEntry{
		InstanceNumber: instanceNumber,
		Priority:       priority,
	}, nil
}
