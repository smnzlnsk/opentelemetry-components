package evaluators

import (
	"math/rand/v2"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
)

// dynamicRandomEvaluator implements interfaces.Evaluator for dynamic random values
type dynamicRandomEvaluator struct {
	identifier string
}

func NewDynamicRandomEvaluator(identifier string) domain.Evaluator {
	return &dynamicRandomEvaluator{
		identifier: identifier,
	}
}

func (d *dynamicRandomEvaluator) Identifier() string {
	return d.identifier
}

func (d *dynamicRandomEvaluator) Evaluate(factor float64, params map[string]interface{}) float64 {
	// Generate a new random value on each evaluation
	return rand.Float64() * factor
}
