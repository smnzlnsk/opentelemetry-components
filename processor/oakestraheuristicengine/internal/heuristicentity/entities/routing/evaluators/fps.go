package evaluators

import (
	"fmt"
	"math"

	"github.com/smnzlnsk/opentelemetry-components/pkg/notification"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
)

type FpsEvaluator struct {
	identifier string

	alertCondition notification.Function
	routeCondition notification.Function
}

func NewFpsEvaluator(identifier string) domain.Evaluator {
	var alertCondition notification.Function
	var routeCondition notification.Function

	alertCondition = func(results map[string]interface{}, instances []int) (bool, error) {
		fmt.Println("triggering alert condition fps", instances)
		return true, nil
		/*for _, instance := range instances {
			prefix := fmt.Sprintf("job.instance.%d", instance) // Create current job instance prefix
			in_0, ok := results[fmt.Sprintf("%s.service_fps(0){in}", prefix)].(float64)
			if !ok {
				return false, fmt.Errorf("service_fps(0){in} not found")
			}
			in_4, ok := results[fmt.Sprintf("%s.service_fps(4){in}", prefix)].(float64)
			if !ok {
				return false, fmt.Errorf("service_fps(4){in} not found")
			}

			// If performance dropped by more than 10% in the last 5 seconds, return true
			percentageChange_bigger_10 := (in_4-in_0)/math.Max(in_0, in_4)*100 > 10
			if percentageChange_bigger_10 {
				fmt.Println("alert condition fps true", instance)
				return true, nil
			}
		}

		// TODO: Here we could start creating a view on the instance combinations and compare the results between instances

		return false, nil*/
	}

	routeCondition = func(results map[string]interface{}, instances []int) (bool, error) {
		for _, instance := range instances {
			prefix := fmt.Sprintf("job.instance.%d", instance) // Create current job instance prefix
			in_0, ok := results[fmt.Sprintf("%s.service_fps(0){in}", prefix)].(float64)
			if !ok {
				return false, fmt.Errorf("service_fps(0){in} not found")
			}
			in_4, ok := results[fmt.Sprintf("%s.service_fps(4){in}", prefix)].(float64)
			if !ok {
				return false, fmt.Errorf("service_fps(4){in} not found")
			}

			// If performance dropped by more than 5% in the last 5 seconds, return true
			percentageChange_bigger_5 := (in_4-in_0)/math.Max(in_0, in_4)*100 > 5
			if percentageChange_bigger_5 {
				fmt.Println("route condition fps true", instance)
				return true, nil
			}
		}

		// TODO: Here we could start creating a view on the instance combinations and compare the results between instances

		return false, nil
	}

	return &FpsEvaluator{
		identifier:     identifier,
		alertCondition: alertCondition,
		routeCondition: routeCondition,
	}
}

// Evaluate evaluates the priority on a per instance basis
func (e *FpsEvaluator) Evaluate(factor float64, arguments map[string]interface{}) (float64, error) {
	n, ok := arguments["service_fps(0){in}"].(float64)
	if !ok {
		return 0, fmt.Errorf("service_fps(0){in} not found")
	}

	return exponentialDecay(n, 0.02), nil
}

func (e *FpsEvaluator) AlarmCondition() notification.Function {
	return e.alertCondition
}

func (e *FpsEvaluator) RouteCondition() notification.Function {
	return e.routeCondition
}

// unsupported
func (e *FpsEvaluator) ScheduleCondition() notification.Function {
	return nil
}
