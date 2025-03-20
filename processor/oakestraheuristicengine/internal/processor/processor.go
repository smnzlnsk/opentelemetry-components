package processor

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
)

// processor implements interfaces.Processor
type processor struct {
	identifier string
	evaluator  interfaces.Evaluator
}

func NewProcessor(identifier string, evaluator interfaces.Evaluator) interfaces.Processor {
	return &processor{
		identifier: identifier,
		evaluator:  evaluator,
	}
}

func (p *processor) Identifier() string {
	return p.identifier
}

func (p *processor) Evaluator() interfaces.Evaluator {
	return p.evaluator
}

func (p *processor) Process(params map[string]interface{}) float64 {
	return p.evaluator.Evaluate(1, params)
}
