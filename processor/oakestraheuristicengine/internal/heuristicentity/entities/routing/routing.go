package routing

import (
	"context"
	"errors"
	"fmt"

	"github.com/Knetic/govaluate"
	"github.com/smnzlnsk/opentelemetry-components/pkg/evaluation"
	"github.com/smnzlnsk/opentelemetry-components/pkg/job"
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
	// resultHistory  map[string]domain.EvaluationResult
	logger *zap.Logger
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

	// Create the processor notification conditions
	// This will inherenetly mean that the entity has the respective notification interfaces set
	expression, err := govaluate.NewEvaluableExpression("true")
	if err != nil {
		logger.Error("Failed to create evaluable expression", zap.Error(err))
		return nil
	}

	conditions := map[domain.NotificationInterfaceCapability]*govaluate.EvaluableExpression{
		domain.NotificationInterfaceCapability_Alert:    expression,
		domain.NotificationInterfaceCapability_Route:    expression,
		domain.NotificationInterfaceCapability_Schedule: expression,
	}

	// Create and add the processors to the processor store
	processorStore.Add(processor.NewProcessor("fps", fpsEvaluator, conditions))
	processorStore.Add(processor.NewProcessor("closest", closestEvaluator, conditions))
	processorStore.Add(processor.NewProcessor("underutilized", underutilizedEvaluator, conditions))

	return &routingEntity{
		processorStore: processorStore,
		logger:         logger,
		services:       services,
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

		result.Values[instanceName] = instanceValues

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

	for _, capability := range notificationOrder {
		if enabled, exists := conditions[capability]; !exists || !enabled {
			continue
		}

		notificationInterface := r.getNotificationInterface(capability)
		if notificationInterface == nil {
			r.logger.Error("No notification interface found for capability", zap.String("capability", string(capability)))
			continue
		}

		condition := processor.GetNotificationCondition(capability)
		if condition == nil {
			r.logger.Error("No notification condition found for capability", zap.String("capability", string(capability)))
			continue
		}

		conditionalResult, err := r.evaluateCondition(condition, result)
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

func (r *routingEntity) evaluateCondition(condition *govaluate.EvaluableExpression, results evaluation.Result) (bool, error) {
	values := make(map[string]interface{})

	for _, result := range results.Results {
		values[fmt.Sprintf("%s.instance.%d", results.JobName, result.InstanceNumber)] = result.Priority
	}

	conditional, err := condition.Evaluate(values)
	if err != nil {
		r.logger.Error("Failed to evaluate condition", zap.Error(err))
		return false, err
	}

	conditionalResult, ok := conditional.(bool)
	if !ok {
		r.logger.Error("Conditional result is not a boolean", zap.Any("conditional", conditional))
		return false, errors.New("conditional result is not a boolean")
	}

	return conditionalResult, nil
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
