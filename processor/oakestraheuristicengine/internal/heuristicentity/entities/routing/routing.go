package routing

import (
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
	processorStore domain.ProcessorStore
	logger         *zap.Logger
}

func NewRoutingEntity(logger *zap.Logger) domain.HeuristicEntity {
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
	}
}

func (r *routingEntity) Evaluate(processorIdentifier string, appname string, values map[string]interface{}) domain.Evaluation {
	return r.processorStore.Get(processorIdentifier).Process(appname, values)
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
