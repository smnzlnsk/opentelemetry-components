package interfaces

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
)

type MeasureFactory interface {
	CreateMeasure(measureType types.MeasureType) MeasureNotifier
}

type MeasureRegistry interface {
	Register(measure MeasureNotifier)
	Get(name types.MeasureType) MeasureNotifier
}

type MeasureNotifier interface {
	Notify() error
	Type() types.MeasureType
}
