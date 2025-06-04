package evaluators

import (
	"github.com/smnzlnsk/opentelemetry-components/pkg/notification"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
)

// closestEvaluator implements domain.Evaluator for the closest service instance
type closestEvaluator struct {
	identifier     string
	alertCondition notification.Function
	routeCondition notification.Function
}

// NewClosestEvaluator creates a new closest evaluator
func NewClosestEvaluator(identifier string) domain.Evaluator {
	var alertCondition notification.Function
	var routeCondition notification.Function

	// TODO: Implement the closest evaluator
	// Then define the alert, route and schedule conditions here
	// We do not want to return these function defintions from the respective functions below,
	// as they would lead to more garbage collection pressure
	alertCondition = nil
	routeCondition = nil

	return &closestEvaluator{
		identifier:     identifier,
		alertCondition: alertCondition,
		routeCondition: routeCondition,
	}
}

var _ domain.Evaluator = &closestEvaluator{}

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

func (e *closestEvaluator) AlarmCondition() notification.Function {
	return e.alertCondition
}

func (e *closestEvaluator) RouteCondition() notification.Function {
	return e.routeCondition
}

// unsupported
func (e *closestEvaluator) ScheduleCondition() notification.Function {
	return nil
}
