package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
)

type AlertConfig struct {
	Threshold float64 `mapstructure:"threshold"`
	Message   string  `mapstructure:"message"`
	Endpoint  string  `mapstructure:"endpoint"`
}

var _ domain.NotificationInterface = (*alertNotifier)(nil)

type alertNotifier struct {
	host       string
	port       int
	endpoint   string
	capability domain.NotificationInterfaceCapability
}

func (a *alertNotifier) Notify(_ interface{}) error {
	jsonData := map[string]interface{}{
		"alert":   "true",
		"message": "test",
	}

	data, err := json.Marshal(jsonData)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("http://%s:%d%s", a.host, a.port, a.endpoint),
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

func (a *alertNotifier) Type() domain.NotificationInterfaceCapability {
	return domain.NotificationInterfaceCapability_Alert
}
