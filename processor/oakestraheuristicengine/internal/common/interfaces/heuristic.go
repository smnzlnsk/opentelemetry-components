package interfaces

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/processor"
)

type HeuristicEntityFactory interface {
	CreateHeuristicEntity(heuristicType types.HeuristicType) (HeuristicEntity, error)
}

type HeuristicEntity interface {
	Processors() map[string]processor.HeuristicProcessor
	AddProcessor(identifier string, processor processor.HeuristicProcessor)
	EvaluatePolicy() error
	Start() error
	Shutdown() error
}
