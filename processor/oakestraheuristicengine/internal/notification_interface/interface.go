package notification_interface

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/logger"
)

type notificationInterface[T any] struct {
	capability domain.NotificationInterfaceCapability
	host       string
	port       int
	endpoint   string
}

var _ domain.NotificationInterface[any] = (*notificationInterface[any])(nil)

func (n *notificationInterface[T]) String() string {
	return fmt.Sprintf("http://%s:%d%s", n.host, n.port, n.endpoint)
}

func (n *notificationInterface[T]) Notify(notification T) error {
	jsonData := map[string]interface{}{
		"type":         n.capability.String(),
		"notification": notification,
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

	var responseStatus string
	if err != nil {
		responseStatus = fmt.Sprintf("error: %s", err.Error())
	} else {
		defer resp.Body.Close()
		responseStatus = resp.Status
	}

	if logger.IsCSVLoggerInitialized() {
		notificationData, _ := json.Marshal(notification)
		logError := logger.LogNotification(n.capability.String(), n.host, responseStatus, string(notificationData))
		if logError != nil {
			fmt.Println("error logging notification", logError)
		}
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send %s: %s", n.capability, resp.Status)
	}

	fmt.Println("notification sent", n.capability, notification)

	return nil
}

func (n *notificationInterface[T]) Type() domain.NotificationInterfaceCapability {
	return n.capability
}
