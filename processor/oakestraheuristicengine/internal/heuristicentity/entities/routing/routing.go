package routing

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/smnzlnsk/opentelemetry-components/pkg/evaluation"
	"github.com/smnzlnsk/opentelemetry-components/pkg/job"
	"github.com/smnzlnsk/opentelemetry-components/pkg/notification"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/heuristicentity/entities/routing/evaluators"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/processor"
	"go.uber.org/zap"
)

type routingEntity struct {
	services       domain.Services
	processorStore domain.ProcessorStore
	alert          domain.NotificationInterface[any]
	route          domain.NotificationInterface[any]
	schedule       domain.NotificationInterface[any]
	prevResults    map[resultHistoryKey]float64
	logger         *zap.Logger
}

type resultHistoryKey struct {
	jobName        string
	instanceNumber int
	ipType         string
}

var notificationOrder = []domain.NotificationInterfaceCapability{
	domain.NotificationInterfaceCapability_Alert,
	domain.NotificationInterfaceCapability_Route,
	domain.NotificationInterfaceCapability_Schedule,
}

func NewRoutingEntity(services domain.Services, logger *zap.Logger) domain.HeuristicEntity {
	processorStore := processor.NewStore()

	// TODO: Add processors here

	// First initialize the processors evaluator
	fpsEvaluator := evaluators.NewFpsEvaluator("fps")
	closestEvaluator := evaluators.NewClosestEvaluator("closest")
	underutilizedEvaluator := evaluators.NewUnderutilizedEvaluator("underutilized")

	// Create and add the processors to the processor store
	processorStore.Add(processor.NewProcessor("fps", fpsEvaluator))
	processorStore.Add(processor.NewProcessor("closest", closestEvaluator))
	processorStore.Add(processor.NewProcessor("underutilized", underutilizedEvaluator))

	return &routingEntity{
		processorStore: processorStore,
		logger:         logger,
		services:       services,
		prevResults:    make(map[resultHistoryKey]float64),
		// notification interfaces are set later
	}
}

// Evaluate evaluates the routing entity
// arguments:
// - processorIdentifier: the identifier of the processor to evaluate
// - first positional argument: jobRequest [domain.JobRequest]
func (r *routingEntity) Evaluate(arguments ...interface{}) error {
	if len(arguments) == 0 {
		r.logger.Error("No arguments provided to Evaluate")
		return errors.New("no arguments provided to Evaluate")
	}

	if len(arguments) < 2 {
		r.logger.Error("Not enough arguments provided to Evaluate")
		return errors.New("not enough arguments provided to Evaluate")
	}

	processorIdentifier, ok := arguments[0].(string)
	if !ok {
		r.logger.Error("First argument is not a string", zap.Any("actual_type", fmt.Sprintf("%T", arguments[0])))
		return errors.New("first argument is not a string")
	}

	jobRequest, ok := arguments[1].(job.Request)
	if !ok {
		r.logger.Error("Second argument is not a JobRequest", zap.Any("actual_type", fmt.Sprintf("%T", arguments[1])))
		return errors.New("second argument is not a JobRequest")
	}
	jobName := jobRequest.JobData.JobName
	instances := jobRequest.JobData.ServiceInstanceList

	result := evaluation.Result{
		JobName: jobName,
		Values:  make(map[string]map[string]interface{}),
		Results: make([]evaluation.Entry, 0, len(instances)),
	}

	// Get metrics for evaluation - metricsService handles memory vs persistent logic internally
	values, err := r.services.GetMetricsService().GetJobMetricsAsMap(
		context.Background(),
		jobName,
	)
	if err != nil {
		r.logger.Error("Failed to get job metrics", zap.Error(err))
		return err
	}

	processor := r.processorStore.Get(processorIdentifier)

	for i := range instances {
		instanceName := fmt.Sprintf("%s.instance.%d", jobName, instances[i].InstanceNumber)
		instanceValues := values.InstanceMetricsForEvaluation(instanceName)

		result.Values[fmt.Sprintf("job.instance.%d", instances[i].InstanceNumber)] = instanceValues

		evalResult, err := processor.Process(instances[i].InstanceNumber, 1, instanceValues)
		if err != nil {
			r.logger.Error("Failed to process instance", zap.Error(err))
			return err
		}
		evalResult.IpType = processorIdentifier

		result.Results = append(
			result.Results,
			evalResult,
		)
	}

	// Now that we have the results, we can evaluate the notification conditions
	// This will trigger a notification about a job, where necessary
	conditions := processor.GetCapabilities()

	vs := make(map[string]interface{})
	is := []int{}
	once := sync.Once{}
	doEvaluate := false

	// Check if we have any conditions to evaluate
	// If we do, we need to flatten the result values for evaluation
	for _, enabled := range conditions {
		if !enabled {
			continue
		}
		once.Do(func() {
			doEvaluate = true
		})
		// Flatten the result values for evaluation
		for _, result := range result.Results {
			current := result.Priority
			is = append(is, result.InstanceNumber)
			previous := r.prevResults[resultHistoryKey{
				jobName:        jobName,
				instanceNumber: result.InstanceNumber,
				ipType:         result.IpType,
			}]
			delta := math.Abs(current - previous)
			vs[fmt.Sprintf("job.instance.%d.result{%s}", result.InstanceNumber, "current")] = current
			vs[fmt.Sprintf("job.instance.%d.result{%s}", result.InstanceNumber, "previous")] = previous
			vs[fmt.Sprintf("job.instance.%d.result{%s}", result.InstanceNumber, "delta")] = delta
		}

		for name, metrics := range result.Values {
			for metricName, metricValue := range metrics {
				vs[fmt.Sprintf("%s.%s", name, metricName)] = metricValue
			}
		}
		break
	}

	// If we don't have any conditions to evaluate, we can return early
	if !doEvaluate {
		return nil
	}

	for _, capability := range notificationOrder {
		if enabled, exists := conditions[capability]; !exists || !enabled {
			continue
		}

		notificationInterface := r.getNotificationInterface(capability)
		if notificationInterface == nil {
			r.logger.Error("No notification interface found for capability", zap.String("capability", string(capability)))
			continue
		}

		condition := processor.GetNotificationFunction(capability)
		if condition == nil {
			r.logger.Error("No notification condition found for capability", zap.String("capability", string(capability)))
			continue
		}

		conditionalResult, err := r.evaluateCondition(condition, vs, is)
		if err != nil {
			r.logger.Error("Failed to evaluate condition", zap.Error(err))
			return err
		}

		if conditionalResult {
			notification := result
			notification.Values = nil
			notificationInterface.Notify(notification)
			break
		}
	}

	// Update the result history
	for _, resultEntry := range result.Results {
		r.prevResults[resultHistoryKey{
			jobName:        result.JobName,
			instanceNumber: resultEntry.InstanceNumber,
			ipType:         resultEntry.IpType,
		}] = resultEntry.Priority
	}

	return nil
}

func (r *routingEntity) getNotificationInterface(capability domain.NotificationInterfaceCapability) domain.NotificationInterface[any] {
	switch capability {
	case domain.NotificationInterfaceCapability_Alert:
		return r.alert
	case domain.NotificationInterfaceCapability_Route:
		return r.route
	case domain.NotificationInterfaceCapability_Schedule:
		return r.schedule
	default:
		return nil
	}
}

func (r *routingEntity) evaluateCondition(conditionEval notification.Function, values map[string]interface{}, instances []int) (bool, error) {
	conditional, err := conditionEval(values, instances)
	if err != nil {
		r.logger.Error("Failed to evaluate condition", zap.Error(err))
		return false, err
	}

	return conditional, nil
}

func (r *routingEntity) Start() error {
	r.logger.Info("Starting routing entity")
	return nil
}

func (r *routingEntity) Shutdown() error {
	r.logger.Info("Shutting down routing entity")
	return nil
}

func (r *routingEntity) AddProcessor(processor domain.Processor) {
	if exists := r.processorStore.Get(processor.Identifier()); exists != nil {
		r.logger.Error("Processor already exists", zap.String("identifier", processor.Identifier()))
		return
	}
	r.processorStore.Add(processor)
}

func (r *routingEntity) Processors() map[string]domain.Processor {
	return r.processorStore.GetAll()
}

func (r *routingEntity) SetAlert(alert domain.NotificationInterface[any]) error {
	if r.alert != nil {
		return errors.New("alert already set")
	}
	r.alert = alert
	return nil
}

func (r *routingEntity) SetRoute(route domain.NotificationInterface[any]) error {
	if r.route != nil {
		return errors.New("route already set")
	}
	r.route = route
	return nil
}

func (r *routingEntity) SetSchedule(schedule domain.NotificationInterface[any]) error {
	if r.schedule != nil {
		return errors.New("schedule already set")
	}
	r.schedule = schedule
	return nil
}

func (r *routingEntity) Alert() domain.NotificationInterface[any] {
	return r.alert
}

func (r *routingEntity) Route() domain.NotificationInterface[any] {
	return r.route
}

func (r *routingEntity) Schedule() domain.NotificationInterface[any] {
	return r.schedule
}
