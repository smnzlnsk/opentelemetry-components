package schedule

import (
	"fmt"
	"net"
	"os"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/constants"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
)

type ScheduleConfig struct {
	Threshold float64 `mapstructure:"threshold"`
	Message   string  `mapstructure:"message"`
	Endpoint  string  `mapstructure:"endpoint"`
}

var _ interfaces.NotificationInterface = (*scheduleNotifier)(nil)

type scheduleNotifier struct {
	host     string
	port     int
	endpoint string
}

func (s *scheduleNotifier) Notify() error {
	host, port := os.Getenv("SCHEDULE_NOTIFIER_HOST"), os.Getenv("SCHEDULE_NOTIFIER_PORT")
	if host == "" || port == "" {
		return fmt.Errorf("SCHEDULE_NOTIFIER_HOST and SCHEDULE_NOTIFIER_PORT must be set")
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return err
	}
	defer conn.Close()

	return nil
}

func (s *scheduleNotifier) Type() types.NotificationInterfaceCapability {
	return constants.NotificationInterfaceCapability_Schedule
}
