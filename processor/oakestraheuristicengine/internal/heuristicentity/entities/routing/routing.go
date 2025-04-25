package routing

import (
	"context"
	"fmt"
	"math/rand/v2"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/processor"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/wpt"
	"go.uber.org/zap"
)

// dynamicRandomEvaluator implements interfaces.Evaluator for dynamic random values
type dynamicRandomEvaluator struct {
	identifier string
}

func newDynamicRandomEvaluator(identifier string) domain.Evaluator {
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

type routingEntity struct {
	services       domain.Services
	processorStore domain.ProcessorStore
	logger         *zap.Logger
}

func NewRoutingEntity(services domain.Services, logger *zap.Logger) domain.HeuristicEntity {
	processorStore := processor.NewStore()

	// TODO: Add processors here

	// Round Robin Processor
	// Default Processor is also Round Robin
	builder := wpt.NewBuilder("true", 1, 0)
	builder.Left("true", 0.5, 0)
	builder.Right("false", 0, 0.5)
	rrTree := builder.BuildTree("rr-tree", 1.0)
	processorStore.Add(processor.NewProcessor("RR", rrTree))
	processorStore.Add(processor.NewProcessor("default", rrTree))

	// Random Processor (static - only generates random value at initialization)
	builder = wpt.NewBuilder("true", rand.Float64(), 0)
	processorStore.Add(processor.NewProcessor("static-random", builder.BuildTree("static-random-tree", 1.0)))

	// Dynamic Random Processor (generates new random value on each evaluation)
	dynamicRandomTree := newDynamicRandomEvaluator("dynamic-random-tree")
	processorStore.Add(processor.NewProcessor("random", dynamicRandomTree))

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
func (r *routingEntity) Evaluate(processorIdentifier string, arguments ...interface{}) domain.EvaluationResult {
	if len(arguments) == 0 {
		r.logger.Error("No arguments provided to Evaluate")
		return domain.EvaluationResult{JobName: "unknown", Values: make(map[string]map[string]interface{})}
	}

	jobRequest, ok := arguments[0].(domain.JobRequest)
	if !ok {
		r.logger.Error("First argument is not a JobRequest", zap.Any("actual_type", fmt.Sprintf("%T", arguments[0])))
		return domain.EvaluationResult{JobName: "unknown", Values: make(map[string]map[string]interface{})}
	}
	jobName := jobRequest.JobData.JobName
	instances := jobRequest.JobData.ServiceInstanceList

	result := domain.EvaluationResult{
		JobName: jobName,
		Values:  make(map[string]map[string]interface{}),
		Results: make([]domain.EvaluationEntry, len(instances)),
	}

	values, err := r.services.GetMetricsService().GetJobMetricsAsMap(
		context.Background(),
		jobName,
	)
	if err != nil {
		r.logger.Error("Failed to get job metrics", zap.Error(err))
		return result
	}

	for i := range instances {
		instanceName := fmt.Sprintf("%s.instance.%d", jobName, instances[i].InstanceNumber)
		instanceValues := values.InstanceMetricsForEvaluation(instanceName)

		result.Values[instanceName] = instanceValues

		result.Results = append(result.Results, domain.EvaluationEntry{
			InstanceNumber: instances[i].InstanceNumber,
			IpType:         jobRequest.IpType,
			Priority:       r.processorStore.Get(processorIdentifier).Evaluator().Evaluate(1, instanceValues),
		})
	}

	return result
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
