package route

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
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
	capability domain.NotificationInterfaceCapability
}

var _ domain.NotificationInterface = (*routeNotifier)(nil)

func (r *routeNotifier) Notify(notification interface{}) error {
	jobData := notification.(domain.Job)

	data, err := json.Marshal(jobData)
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

func (r *routeNotifier) Type() domain.NotificationInterfaceCapability {
	return domain.NotificationInterfaceCapability_Route
}
