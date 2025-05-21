package evaluators

import "github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"

type roundRobinEvaluator struct {
	identifier string
}

func NewRoundRobinEvaluator(identifier string) domain.Evaluator {
	return &roundRobinEvaluator{
		identifier: identifier,
	}
}

func (e *roundRobinEvaluator) Identifier() string {
	return e.identifier
}

func (e *roundRobinEvaluator) Evaluate(factor float64, params map[string]interface{}) (float64, error) {
	// TODO: Implement the round robin evaluator
	// This evaluator should evaluate the round robin value
	// The round robin value is a value that is incremented on each call of the service instance
	// and is used to determine the next service instance to route to
	// As the usage of the service instances is decoupled from the routing entity,
	// we will leave it up to the node's network manager to decide, who to route to
	return 1 * factor, nil
}
