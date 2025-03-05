package route

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

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
	host       string
	port       int
	endpoint   string
	capability types.NotificationInterfaceCapability
}

var _ interfaces.NotificationInterface = (*routeNotifier)(nil)

func (r *routeNotifier) Notify() error {
	jsonData := map[string]interface{}{
		"route":   "true",
		"message": "test",
	}

	data, err := json.Marshal(jsonData)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("http://%s:%d%s", r.host, r.port, r.endpoint),
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

func (r *routeNotifier) Type() types.NotificationInterfaceCapability {
	return constants.NotificationInterfaceCapability_Route
}
