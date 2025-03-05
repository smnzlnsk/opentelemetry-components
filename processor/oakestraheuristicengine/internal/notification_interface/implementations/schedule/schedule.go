package schedule

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
)

type ScheduleConfig struct {
	Threshold float64 `mapstructure:"threshold"`
	Message   string  `mapstructure:"message"`
	Endpoint  string  `mapstructure:"endpoint"`
}

// scheduleNotifier implements the NotificationInterface interface
type scheduleNotifier struct {
	host       string
	port       int
	endpoint   string
	capability types.NotificationInterfaceCapability
}

var _ interfaces.NotificationInterface = (*scheduleNotifier)(nil)

func (s *scheduleNotifier) Notify() error {
	jsonData := map[string]interface{}{
		"route":   "true",
		"message": "test",
	}

	data, err := json.Marshal(jsonData)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("http://%s:%d%s", s.host, s.port, s.endpoint),
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send alert: %s", resp.Status)
	}

	return nil
}

func (s *scheduleNotifier) Type() types.NotificationInterfaceCapability {
	return s.capability
}
