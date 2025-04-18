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

func (p *processor) Process(jobname string, params map[string]interface{}) domain.Evaluation {
	return domain.Evaluation{
		JobName: jobname,
		Entries: []domain.EvaluationEntry{
			{InstanceNumber: 1, Priority: p.evaluator.Evaluate(1, params)},
		},
	}
}
