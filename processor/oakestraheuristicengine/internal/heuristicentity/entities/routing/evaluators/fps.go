package evaluators

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
)

type FpsEvaluator struct {
	identifier string
}

func NewFpsEvaluator(identifier string) domain.Evaluator {
	return &FpsEvaluator{
		identifier: identifier,
	}
}

func (e *FpsEvaluator) Evaluate(factor float64, arguments map[string]interface{}) (float64, error) {

	return 0, nil
}
