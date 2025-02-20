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

// routeNotifier implements the NotificationInterface interface
type routeNotifier struct {
	host     string
	port     int
	endpoint string
}

var _ interfaces.NotificationInterface = (*routeNotifier)(nil)

func (r *routeNotifier) Notify() error {
	return nil
}

func (r *routeNotifier) Type() types.NotificationInterfaceCapability {
	return constants.NotificationInterfaceCapability_Route
}
