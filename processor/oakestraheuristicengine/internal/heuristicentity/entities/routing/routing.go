package routing

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/measure"
	"go.uber.org/zap"
)

type routingEntity struct {
	logger *zap.Logger
}

func NewRoutingEntity(logger *zap.Logger) *routingEntity {
	return &routingEntity{
		logger: logger,
	}
}

func (r *routingEntity) EvaluatePolicy() map[string]measure.Measure {
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
