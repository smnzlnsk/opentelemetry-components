package alert

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/constants"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
)

type AlertConfig struct {
	Threshold float64 `mapstructure:"threshold"`
	Message   string  `mapstructure:"message"`
	Endpoint  string  `mapstructure:"endpoint"`
}

var _ interfaces.NotificationInterface = (*alertNotifier)(nil)

type alertNotifier struct {
	host     string
	port     int
	endpoint string
}

func (a *alertNotifier) Notify() error {
	return nil
}

func (a *alertNotifier) Type() types.NotificationInterfaceCapability {
	return constants.NotificationInterfaceCapability_Alert
}
