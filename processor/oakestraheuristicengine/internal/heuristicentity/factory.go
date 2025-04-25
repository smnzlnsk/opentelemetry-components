package heuristicentity

import (
	"fmt"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/heuristicentity/entities/routing"
	"go.uber.org/zap"
)

type heuristicEntityFactory struct {
	logger *zap.Logger
}

func NewHeuristicEntityFactory(logger *zap.Logger) domain.HeuristicEntityFactory {
	return &heuristicEntityFactory{logger: logger}
}

func (f *heuristicEntityFactory) CreateHeuristicEntity(heuristicType domain.HeuristicType, services domain.Services) (domain.HeuristicEntity, error) {
	switch heuristicType {
	case domain.RoutingEntity:
		return routing.NewRoutingEntity(services, f.logger), nil
	default:
		return nil, fmt.Errorf("heuristic type %s not found", heuristicType)
	}
}
