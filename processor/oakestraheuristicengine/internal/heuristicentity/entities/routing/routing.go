package routing

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/smnzlnsk/opentelemetry-components/internal/shared/evaluation"
	"github.com/smnzlnsk/opentelemetry-components/internal/shared/job"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/heuristicentity/entities/routing/evaluators"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/processor"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/wpt"
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

func NewRoutingEntity(services domain.Services, logger *zap.Logger) domain.HeuristicEntity {
	processorStore := processor.NewStore()

	// TODO: Add processors here

	// Round Robin Processor
	// Default Processor is also Round Robin
	builder := wpt.NewBuilder("true", 1, 0)
	builder.Left("true", 0.5, 0)
	builder.Right("false", 0, 0.5)
	_ = builder.BuildTree("rr-tree", 1.0)

	// Random Processor (static - only generates random value at initialization)
	builder = wpt.NewBuilder("true", rand.Float64(), 0)
	processorStore.Add(processor.NewProcessor("static-random", builder.BuildTree("static-random-tree", 1.0), nil))

	// Dynamic Random Processor (generates new random value on each evaluation)
	dynamicRandomTree := evaluators.NewDynamicRandomEvaluator("dynamic-random")
	closestEvaluator := evaluators.NewClosestEvaluator("closest")
	underutilizedEvaluator := evaluators.NewUnderutilizedEvaluator("underutilized")

	processorStore.Add(processor.NewProcessor("random", dynamicRandomTree, nil))
	processorStore.Add(processor.NewProcessor("RR", dynamicRandomTree, nil))
	processorStore.Add(processor.NewProcessor("closest", closestEvaluator, nil))
	processorStore.Add(processor.NewProcessor("underutilized", underutilizedEvaluator, nil))

	return &routingEntity{
		processorStore: processorStore,
		logger:         logger,
		services:       services,
	}
}

// Evaluate evaluates the routing entity
// arguments:
// - processorIdentifier: the identifier of the processor to evaluate
// - first positional argument: jobRequest [domain.JobRequest]
func (r *routingEntity) Evaluate(processorIdentifier string, arguments ...interface{}) error {
	if len(arguments) == 0 {
		r.logger.Error("No arguments provided to Evaluate")
		return errors.New("no arguments provided to Evaluate")
	}

	jobRequest, ok := arguments[0].(job.Request)
	if !ok {
		r.logger.Error("First argument is not a JobRequest", zap.Any("actual_type", fmt.Sprintf("%T", arguments[0])))
		return errors.New("first argument is not a JobRequest")
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

	for i := range instances {
		instanceName := fmt.Sprintf("%s.instance.%d", jobName, instances[i].InstanceNumber)
		instanceValues := values.InstanceMetricsForEvaluation(instanceName)

		result.Values[instanceName] = instanceValues

		evalResult, err := r.processorStore.Get(processorIdentifier).Process(instances[i].InstanceNumber, 1, instanceValues)
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

	return nil
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
