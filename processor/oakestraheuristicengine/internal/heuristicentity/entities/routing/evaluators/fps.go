package evaluators

import (
	"fmt"
	"math"

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
	n, ok := arguments["service_fps(0){in}"].(float64)
	if !ok {
		return 0, fmt.Errorf("service_fps(0){in} not found")
	}

	return exponentialDecay(n, 0.02), nil
}

func exponentialDecay(n, k float64) float64 {
	return 1 - math.Exp(-k*n)
}
