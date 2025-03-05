package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

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
	host       string
	port       int
	endpoint   string
	capability types.NotificationInterfaceCapability
}

func (a *alertNotifier) Notify() error {
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

func (a *alertNotifier) Type() types.NotificationInterfaceCapability {
	return constants.NotificationInterfaceCapability_Alert
}
