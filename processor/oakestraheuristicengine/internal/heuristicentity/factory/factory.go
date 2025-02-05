package factory

import (
	"fmt"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/heuristicentity/entities/routing"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/heuristicentity/interfaces"
	"go.uber.org/zap"
)

type heuristicEntityFactory struct {
	logger *zap.Logger
}

func NewHeuristicEntityFactory(logger *zap.Logger) HeuristicEntityFactory {
	return &heuristicEntityFactory{logger: logger}
}

func (f *heuristicEntityFactory) CreateHeuristicEntity(heuristicType HeuristicType) (interfaces.HeuristicEntity, error) {
	switch heuristicType {
	case RoutingEntity:
		return routing.NewRoutingEntity(f.logger), nil
	default:
		return nil, fmt.Errorf("heuristic type %s not found", heuristicType)
	}
}
