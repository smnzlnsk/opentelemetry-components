package evaluators

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
)

type underutilizedEvaluator struct {
	identifier string
}

func NewUnderutilizedEvaluator(identifier string) domain.Evaluator {
	return &underutilizedEvaluator{
		identifier: identifier,
	}
}

func (e *underutilizedEvaluator) Evaluate(factor float64, params map[string]interface{}) float64 {
	// TODO: Implement the underutilized evaluator
	// This evaluator should evaluate the underutilization factor of the service instances

	return 1 * factor
}
