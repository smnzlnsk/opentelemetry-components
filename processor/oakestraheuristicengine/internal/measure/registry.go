package measure

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
	"go.uber.org/zap"
)

type measureRegistry struct {
	logger   *zap.Logger
	measures map[types.MeasureType]interfaces.MeasureNotifier
}

func NewMeasureRegistry(logger *zap.Logger) interfaces.MeasureRegistry {
	return &measureRegistry{
		logger:   logger,
		measures: make(map[types.MeasureType]interfaces.MeasureNotifier),
	}
}

func (r *measureRegistry) Register(measure interfaces.MeasureNotifier) {
	r.measures[measure.Type()] = measure
}

func (r *measureRegistry) Get(name types.MeasureType) interfaces.MeasureNotifier {
	return r.measures[name]
}
