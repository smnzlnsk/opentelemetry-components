package routing

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/processor"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/wpt"
	"go.uber.org/zap"
)

type routingEntity struct {
	processorStore interfaces.ProcessorStore
	logger         *zap.Logger
}

func NewRoutingEntity(logger *zap.Logger) interfaces.HeuristicEntity {
	processorStore := processor.NewStore()

	// TODO: Add processors here
	builder := wpt.NewBuilder("true", 1, 0)
	builder.Left("true", 0.5, 0)
	builder.Right("false", 0, 0.5)
	processorStore.Add(processor.NewProcessor("def", builder.BuildTree("rr-tree", 1.0)))

	return &routingEntity{
		processorStore: processorStore,
		logger:         logger,
	}
}

func (r *routingEntity) Evaluate(processorIdentifier string, values map[string]interface{}) float64 {
	return r.processorStore.Get(processorIdentifier).Process(values)
}

func (r *routingEntity) Start() error {
	r.logger.Info("Starting routing entity")
	return nil
}

func (r *routingEntity) Shutdown() error {
	r.logger.Info("Shutting down routing entity")
	return nil
}

func (r *routingEntity) AddProcessor(processor interfaces.Processor) {
	if exists := r.processorStore.Get(processor.Identifier()); exists != nil {
		r.logger.Error("Processor already exists", zap.String("identifier", processor.Identifier()))
		return
	}
	r.processorStore.Add(processor)
}

func (r *routingEntity) Processors() map[string]interfaces.Processor {
	return r.processorStore.GetAll()
}
