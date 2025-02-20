package routing

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"go.uber.org/zap"
)

type routingEntity struct {
	processors map[string]interfaces.HeuristicProcessor
	logger     *zap.Logger
}

func NewRoutingEntity(logger *zap.Logger) interfaces.HeuristicEntity {
	processors := make(map[string]interfaces.HeuristicProcessor)

	// TODO: Add processors here

	return &routingEntity{
		processors: processors,
		logger:     logger,
	}
}

func (r *routingEntity) Evaluate(values map[string]interface{}) map[string]interface{} {
	r.logger.Info("Evaluating routing policy")
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

func (r *routingEntity) AddProcessor(identifier string, processor interfaces.HeuristicProcessor) {
	if _, ok := r.processors[identifier]; ok {
		r.logger.Error("Processor already exists", zap.String("identifier", identifier))
		return
	}
	r.processors[identifier] = processor
}

func (r *routingEntity) Processors() map[string]interfaces.HeuristicProcessor {
	return r.processors
}
