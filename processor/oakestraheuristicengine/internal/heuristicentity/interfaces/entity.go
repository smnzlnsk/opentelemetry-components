package interfaces

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/measure"
)

type HeuristicEntity interface {
	EvaluatePolicy() map[string]measure.Measure
	Start() error
	Shutdown() error
}
