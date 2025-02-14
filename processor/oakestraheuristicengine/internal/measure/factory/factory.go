package factory

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/constants"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/measure/implementations/alert"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/measure/implementations/route"
	"go.uber.org/zap"
)

type measureFactory struct {
	logger   *zap.Logger
	measures map[types.MeasureType]interfaces.MeasureNotifier
}

func NewMeasureFactory(logger *zap.Logger) interfaces.MeasureFactory {
	return &measureFactory{
		logger:   logger,
		measures: make(map[types.MeasureType]interfaces.MeasureNotifier),
	}
}

func (f *measureFactory) CreateMeasure(measureType types.MeasureType) interfaces.MeasureNotifier {

	var notifier interfaces.MeasureNotifier
	var err error

	switch measureType {

	case constants.MeasureTypeRoute:
		notifier, err = route.NewRouteNotifier(route.RouteConfig{})
		if err != nil {
			return nil
		}

	case constants.MeasureTypeAlert:
		notifier, err = alert.NewAlertNotifier(alert.AlertConfig{})
		if err != nil {
			return nil
		}

	default:
		return nil
	}

	return notifier
}
