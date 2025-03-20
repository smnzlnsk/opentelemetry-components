package interfaces

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
)

type HeuristicEntityFactory interface {
	CreateHeuristicEntity(heuristicType types.HeuristicType) (HeuristicEntity, error)
}

type HeuristicEntity interface {
	Processors() map[string]Processor
	AddProcessor(processor Processor)
	Evaluate(processorIdentifier string, values map[string]interface{}) float64
	Start() error
	Shutdown() error
}
