package logger

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
	"time"
)

type csvLogger struct {
	file   *os.File
	writer *csv.Writer
	mutex  sync.Mutex
}

var globalCSVLogger *csvLogger
var loggerOnce sync.Once

// InitCSVLogger initializes the global CSV logger with the specified file path
func InitCSVLogger(filePath string) error {
	var err error
	loggerOnce.Do(func() {
		globalCSVLogger = &csvLogger{}

		// Create or open CSV file
		globalCSVLogger.file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return
		}

		globalCSVLogger.writer = csv.NewWriter(globalCSVLogger.file)

		// Check if file is empty and write headers
		fileInfo, statErr := globalCSVLogger.file.Stat()
		if statErr != nil {
			// create file if it doesn't exist
			globalCSVLogger.file, err = os.Create(filePath)
			if err != nil {
				return
			}
			globalCSVLogger.writer = csv.NewWriter(globalCSVLogger.file)
			return
		}

		if fileInfo.Size() == 0 {
			headers := []string{"timestamp", "type", "target_host", "response_status", "best_instance", "notification_data"}
			err = globalCSVLogger.writer.Write(headers)
			if err != nil {
				return
			}
			globalCSVLogger.writer.Flush()
		}
	})
	return err
}

// LogNotification logs a notification attempt to CSV
func LogNotification(capability string, targetHost string, responseStatus string, notificationData string) error {
	if globalCSVLogger == nil {
		return fmt.Errorf("CSV logger not initialized")
	}

	return globalCSVLogger.logNotification(capability, targetHost, responseStatus, notificationData)
}

func (c *csvLogger) logNotification(capability string, targetHost string, responseStatus string, notificationData string) error {
	if c == nil {
		return fmt.Errorf("CSV logger not initialized")
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Get current timestamp in microseconds
	timestamp := time.Now().UnixMicro()

	record := []string{
		fmt.Sprintf("%d", timestamp),
		capability,
		targetHost,
		responseStatus,
		notificationData,
	}

	err := c.writer.Write(record)
	if err != nil {
		return err
	}

	// Flush immediately for real-time logging
	c.writer.Flush()
	return c.writer.Error()
}

// CloseCSVLogger should be called when shutting down to ensure all data is written
func CloseCSVLogger() error {
	if globalCSVLogger != nil {
		globalCSVLogger.mutex.Lock()
		defer globalCSVLogger.mutex.Unlock()

		if globalCSVLogger.writer != nil {
			globalCSVLogger.writer.Flush()
		}
		if globalCSVLogger.file != nil {
			return globalCSVLogger.file.Close()
		}
	}
	return nil
}

// IsCSVLoggerInitialized returns true if the CSV logger has been initialized
func IsCSVLoggerInitialized() bool {
	return globalCSVLogger != nil
}
