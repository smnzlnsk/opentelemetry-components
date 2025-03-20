package notification_interface

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
)

type notificationInterface struct {
	capability types.NotificationInterfaceCapability
	host       string
	port       int
	endpoint   string
}

func (n *notificationInterface) Notify() error {
	jsonData := map[string]interface{}{
		n.capability.String(): "true",
		"message":             "test",
	}

	data, err := json.Marshal(jsonData)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("http://%s:%d%s", n.host, n.port, n.endpoint),
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send %s: %s", n.capability, resp.Status)
	}

	return nil
}

func (n *notificationInterface) Type() types.NotificationInterfaceCapability {
	return n.capability
}
