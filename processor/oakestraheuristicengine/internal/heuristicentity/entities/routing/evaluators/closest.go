package evaluators

import "github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"

// closestEvaluator implements domain.Evaluator for the closest service instance
type closestEvaluator struct {
	identifier string
}

// NewClosestEvaluator creates a new closest evaluator
func NewClosestEvaluator(identifier string) domain.Evaluator {
	return &closestEvaluator{
		identifier: identifier,
	}
}

func (e *closestEvaluator) Identifier() string {
	return e.identifier
}

func (e *closestEvaluator) Evaluate(factor float64, params map[string]interface{}) (float64, error) {
	// TODO: Implement the closest evaluator
	// This evaluator should evaluate the closest service instance to the current service instance node that is asking
	// Therefore, we will have to include the current node location in the metrics data
	// We should best start by implementing a geo-location receiver in the monitoring-agent
	// enabling us to calculate the rough distance estimate between the nodes of interest
	return 1, nil
}
