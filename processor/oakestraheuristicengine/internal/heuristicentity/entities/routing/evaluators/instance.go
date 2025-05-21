package evaluators

import "github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"

// instanceEvaluator implements domain.Evaluator for instance service IP type
type instanceEvaluator struct {
	identifier string
}

// NewInstanceEvaluator creates a new instance evaluator
// It always returns 1 as the evaluation result,
// because the instance service IP type is distinctly identifying
// the service instance itself, not a set of instances
func NewInstanceEvaluator(identifier string) domain.Evaluator {
	return &instanceEvaluator{
		identifier: identifier,
	}
}

func (e *instanceEvaluator) Identifier() string {
	return e.identifier
}

func (e *instanceEvaluator) Evaluate(factor float64, params map[string]interface{}) (float64, error) {
	return 1, nil
}
