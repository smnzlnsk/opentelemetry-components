package heuristicentity

import (
	"fmt"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/constants"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/heuristicentity/entities/routing"
	"go.uber.org/zap"
)

type heuristicEntityFactory struct {
	logger *zap.Logger
}

func NewHeuristicEntityFactory(logger *zap.Logger) interfaces.HeuristicEntityFactory {
	return &heuristicEntityFactory{logger: logger}
}

func (f *heuristicEntityFactory) CreateHeuristicEntity(heuristicType types.HeuristicType) (interfaces.HeuristicEntity, error) {
	switch heuristicType {
	case constants.RoutingEntity:
		return routing.NewRoutingEntity(f.logger), nil
	default:
		return nil, fmt.Errorf("heuristic type %s not found", heuristicType)
	}
}
