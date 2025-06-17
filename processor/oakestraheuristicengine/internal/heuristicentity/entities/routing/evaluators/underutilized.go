package evaluators

import (
	"fmt"

	"github.com/smnzlnsk/opentelemetry-components/pkg/notification"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
)

type underutilizedEvaluator struct {
	identifier     string
	alertCondition notification.Function
	routeCondition notification.Function
}

func NewUnderutilizedEvaluator(identifier string) domain.Evaluator {
	var alertCondition notification.Function
	var routeCondition notification.Function

	// TODO: Implement the underutilized evaluator
	// Then define the alert, route and schedule conditions here
	// We do not want to return these function defintions from the respective functions below,
	// as they would lead to more garbage collection pressure
	alertCondition = func(results map[string]interface{}, instances []int) (bool, error) {
		/*
			for _, instance := range instances {
				prefix := fmt.Sprintf("job.instance.%d", instance) // Create current job instance prefix

				delta, ok := results[fmt.Sprintf("%s.result{delta}", prefix)].(float64)
				if !ok {
					continue
				}
				if delta > 0.2 {
					return true, nil
				}

				usageTotal, ok := results[fmt.Sprintf("%s.result{current}", prefix)].(float64)
				if !ok {
					continue
				}
				if usageTotal > 0.9 {
					return true, nil
				}
			}
			return false, nil
		*/
		return true, nil
	}

	routeCondition = func(results map[string]interface{}, instances []int) (bool, error) {
		for _, instance := range instances {
			prefix := fmt.Sprintf("job.instance.%d", instance) // Create current job instance prefix
			usageTotal, ok := results[fmt.Sprintf("%s.result{current}", prefix)].(float64)
			if !ok {
				continue
			}
			if usageTotal > 0.5 {
				return true, nil
			}
		}
		return false, nil
	}

	return &underutilizedEvaluator{
		identifier:     identifier,
		alertCondition: alertCondition,
		routeCondition: routeCondition,
	}
}

func (e *underutilizedEvaluator) Evaluate(factor float64, arguments map[string]interface{}) (float64, error) {
	// TODO: Implement the underutilized evaluator
	// This evaluator should evaluate the underutilization factor of the service instances
	cpuUsageUser, ok := arguments["service.cpu.utilisation(0){user}"].(float64)
	if !ok {
		return 0, fmt.Errorf("service.cpu.utilisation(0){user} not found")
	}

	cpuUsageSystem, ok := arguments["service.cpu.utilisation(0){system}"].(float64)
	if !ok {
		return 0, fmt.Errorf("service.cpu.utilisation(0){system} not found")
	}

	memoryUsageSlabReclaimable, ok := arguments["service.memory.utilisation(0){slab_reclaimable}"].(float64)
	if !ok {
		return 0, fmt.Errorf("service.memory.utilisation(0){slab_reclaimable} not found")
	}

	memoryUsageSlabUnreclaimable, ok := arguments["service.memory.utilisation(0){slab_unreclaimable}"].(float64)
	if !ok {
		return 0, fmt.Errorf("service.memory.utilisation(0){slab_unreclaimable} not found")
	}

	memoryUsageUsed, ok := arguments["service.memory.utilisation(0){used}"].(float64)
	if !ok {
		return 0, fmt.Errorf("service.memory.utilisation(0){used} not found")
	}

	memoryUsageTotal := memoryUsageSlabReclaimable + memoryUsageSlabUnreclaimable + memoryUsageUsed
	cpuUsageTotal := cpuUsageUser + cpuUsageSystem

	usageTotal := cpuUsageTotal + memoryUsageTotal

	res := 1 - exponentialDecay(usageTotal, 0.02)
	return res, nil
}

func (e *underutilizedEvaluator) AlarmCondition() notification.Function {
	return e.alertCondition
}

func (e *underutilizedEvaluator) RouteCondition() notification.Function {
	return e.routeCondition
}

func (e *underutilizedEvaluator) ScheduleCondition() notification.Function {
	return nil // unsupported
}
