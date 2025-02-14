package route

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/constants"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
)

type RouteConfig struct {
	Threshold float64 `mapstructure:"threshold"`
	Message   string  `mapstructure:"message"`
	Endpoint  string  `mapstructure:"endpoint"`
}

type routeNotifier struct {
	cfg RouteConfig
}

var _ interfaces.MeasureNotifier = (*routeNotifier)(nil)

func NewRouteNotifier(cfg RouteConfig) (interfaces.MeasureNotifier, error) {
	return &routeNotifier{cfg: cfg}, nil
}

func (r *routeNotifier) Notify() error {
	return nil
}

func (r *routeNotifier) Type() types.MeasureType {
	return constants.MeasureTypeRoute
}
