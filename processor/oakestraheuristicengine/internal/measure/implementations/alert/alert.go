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

var _ interfaces.MeasureNotifier = (*alertNotifier)(nil)

type alertNotifier struct {
	cfg AlertConfig
}

func NewAlertNotifier(cfg AlertConfig) (interfaces.MeasureNotifier, error) {
	return &alertNotifier{cfg: cfg}, nil
}

func (a *alertNotifier) Notify() error {
	return nil
}

func (a *alertNotifier) Type() types.MeasureType {
	return constants.MeasureTypeAlert
}
