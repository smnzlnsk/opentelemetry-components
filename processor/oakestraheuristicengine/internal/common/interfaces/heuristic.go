package interfaces

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
)

type HeuristicEntityFactory interface {
	CreateHeuristicEntity(heuristicType types.HeuristicType) (HeuristicEntity, error)
}

type HeuristicEntity interface {
	Processors() map[string]HeuristicProcessor
	AddProcessor(identifier string, processor HeuristicProcessor)
	Evaluate(values map[string]interface{}) map[string]interface{}
	Start() error
	Shutdown() error
}
