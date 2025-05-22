package unixsocketreceiver // import github.com/smnzlnsk/opentelemetry-components/receiver/unixsocketreceiver
import (
	"bytes"
	"context"
	"net"
	"sync"
	"time"

	"github.com/smnzlnsk/opentelemetry-components/receiver/unixsocketreceiver/internal/dirwatcher"
	"github.com/smnzlnsk/opentelemetry-components/receiver/unixsocketreceiver/internal/socketreader"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

var _ receiver.Metrics = (*unixSocketReceiver)(nil)

type unixSocketReceiver struct {
	config           *Config
	logger           *zap.Logger
	consumer         consumer.Metrics
	host             component.Host
	cancel           context.CancelFunc
	dirWatcher       *dirwatcher.DirectoryWatcher
	connections      map[string]net.Conn
	mu               sync.Mutex
	prometheusParser *PrometheusParser
}

func newUnixSocketReceiver(config *Config, logger *zap.Logger, consumer consumer.Metrics) (*unixSocketReceiver, error) {
	usr := &unixSocketReceiver{
		config:      config,
		logger:      logger,
		consumer:    consumer,
		connections: make(map[string]net.Conn),
	}
	return usr, nil
}

func (usr *unixSocketReceiver) Start(ctx context.Context, host component.Host) error {
	_, usr.cancel = context.WithCancel(ctx)
	usr.host = host
	usr.logger.Info("Starting Unix Socket Receiver")

	parsedInterval, err := time.ParseDuration(usr.config.Interval)
	if err != nil {
		return err
	}

	usr.prometheusParser = NewPrometheusParser(usr.logger)

	usr.dirWatcher = dirwatcher.NewDirectoryWatcher(usr.config.Folder, usr.logger, parsedInterval, usr.connectToUnixSocket)
	usr.dirWatcher.Start()

	return nil
}

func (usr *unixSocketReceiver) Shutdown(_ context.Context) error {
	if usr.cancel != nil {
		usr.cancel()
	}

	// Stop the directory watcher
	if usr.dirWatcher != nil {
		usr.dirWatcher.Stop()
	}

	// Close all active connections
	usr.mu.Lock()
	for socketPath, conn := range usr.connections {
		usr.logger.Info("Closing connection to unix socket", zap.String("socket", socketPath))
		_ = conn.Close()
		delete(usr.connections, socketPath)
	}
	usr.mu.Unlock()

	usr.logger.Info("Shutdown Unix Socket Receiver")
	return nil
}

func (usr *unixSocketReceiver) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	if usr.consumer == nil {
		usr.logger.Error("no next consumer available for unix socket receiver")
		return nil
	}

	err := usr.consumer.ConsumeMetrics(ctx, md)
	if err != nil {
		usr.logger.Error("failed to consume metrics", zap.Error(err))
	}
	return err
}

func (usr *unixSocketReceiver) connectToUnixSocket(socketPath string) {
	usr.logger.Info("Connecting to unix socket", zap.String("socket", socketPath))

	// Check if we're already connected to this socket
	usr.mu.Lock()
	if _, exists := usr.connections[socketPath]; exists {
		usr.logger.Info("Already connected to unix socket", zap.String("socket", socketPath))
		usr.mu.Unlock()
		return
	}
	usr.mu.Unlock()

	// Create a direct connection to the Unix socket
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		usr.logger.Error("Failed to connect to unix socket",
			zap.String("socket", socketPath),
			zap.Error(err))
		return
	}

	// Store the connection
	usr.mu.Lock()
	usr.connections[socketPath] = conn
	usr.mu.Unlock()

	usr.logger.Info("Successfully connected to unix socket",
		zap.String("socket", socketPath))

	// Start receiving metrics in a separate goroutine
	usr.logger.Info("Starting to receive metrics from unix socket", zap.String("socket", socketPath))
	go usr.receiveMetrics(socketPath, conn)
}

func (usr *unixSocketReceiver) receiveMetrics(socketPath string, conn net.Conn) {
	defer func() {
		usr.logger.Info("Closing connection to unix socket", zap.String("socket", socketPath))
		_ = conn.Close()

		usr.mu.Lock()
		delete(usr.connections, socketPath)
		usr.mu.Unlock()

		// Attempt to reconnect after a delay
		go usr.reconnectWithBackoff(socketPath)
	}()

	ctx := context.Background()

	usr.logger.Info("Starting to read from unix socket", zap.String("socket", socketPath))

	// Create both Proto and JSON unmarshalers to try different formats
	protoUnmarshaler := &pmetric.ProtoUnmarshaler{}
	jsonUnmarshaler := &pmetric.JSONUnmarshaler{}

	// Set up a ticker for reading every second
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// Create socket reader
	reader, err := socketreader.NewSocketReader(socketPath, usr.logger)
	if err != nil {
		usr.logger.Error("Failed to create socket reader",
			zap.String("socket", socketPath),
			zap.Error(err))
		return
	}
	defer reader.Close()

	for {
		select {
		case <-ticker.C:
			data, err := reader.Read(ctx)
			if err != nil {
				usr.logger.Error("Failed to read from socket",
					zap.String("socket", socketPath),
					zap.Error(err))
				return
			}

			if len(data) > 0 {
				usr.processReceivedData(ctx, data, socketPath, protoUnmarshaler, jsonUnmarshaler)
			} else {
				usr.logger.Debug("No data received from socket", zap.String("socket", socketPath))
			}
		}
	}
}

// reconnectWithBackoff attempts to reconnect to a socket with exponential backoff
func (usr *unixSocketReceiver) reconnectWithBackoff(socketPath string) {
	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second
	retries := 0

	for {
		// Wait before attempting to reconnect
		time.Sleep(backoff)

		usr.logger.Info("Attempting to reconnect to unix socket",
			zap.String("socket", socketPath),
			zap.Duration("backoff", backoff),
			zap.Int("retry", retries+1))

		// Check if the socket file exists
		if _, err := net.Dial("unix", socketPath); err != nil {
			usr.logger.Debug("Socket not available for reconnection",
				zap.String("socket", socketPath),
				zap.Error(err))

			// Increase backoff with exponential strategy
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}

			retries++
			continue
		}

		// Socket is available, attempt to connect
		usr.connectToUnixSocket(socketPath)
		return
	}
}

// processReceivedData processes the received data regardless of the protocol
func (usr *unixSocketReceiver) processReceivedData(ctx context.Context, data []byte, socketPath string, protoUnmarshaler *pmetric.ProtoUnmarshaler, jsonUnmarshaler *pmetric.JSONUnmarshaler) {
	// Try removing null bytes and leading/trailing whitespace
	cleanData := cleanBuffer(data)
	if len(cleanData) > 0 {
		// Try text-based metrics format (e.g. Prometheus exposition format) first
		if IsPrometheusFormat(cleanData) {
			metrics, err := usr.prometheusParser.ParsePrometheusMetrics(socketPath, cleanData)
			if err != nil {
				usr.logger.Error("Failed to parse Prometheus format metrics",
					zap.String("socket", socketPath),
					zap.Error(err))
			} else {
				usr.processMetrics(ctx, metrics, socketPath)
			}
			return
		}

		// Try to unmarshal with Proto format
		metrics, err := protoUnmarshaler.UnmarshalMetrics(cleanData)
		if err != nil {
			usr.logger.Error("Failed to unmarshal metrics data with Proto format, trying JSON",
				zap.String("socket", socketPath),
				zap.Error(err))

			// Try JSON format as fallback
			metrics, err = jsonUnmarshaler.UnmarshalMetrics(cleanData)
			if err != nil {
				usr.logger.Error("Failed to unmarshal metrics data with both Proto and JSON formats",
					zap.String("socket", socketPath),
					zap.Error(err))
				return
			}
		}

		// Successfully unmarshaled, process the metrics
		usr.processMetrics(ctx, metrics, socketPath)
	} else {
		usr.logger.Debug("No data after cleaning buffer", zap.String("socket", socketPath))
	}
}

// processMetrics processes the unmarshaled metrics
func (usr *unixSocketReceiver) processMetrics(ctx context.Context, metrics pmetric.Metrics, socketPath string) {
	if err := usr.ConsumeMetrics(ctx, metrics); err != nil {
		usr.logger.Error("Failed to consume metrics",
			zap.String("socket", socketPath),
			zap.Error(err))
	}
}

// cleanBuffer removes null bytes and trims whitespace
func cleanBuffer(data []byte) []byte {
	// Remove null bytes
	withoutNulls := bytes.ReplaceAll(data, []byte{0}, []byte{})
	// Trim whitespace
	return bytes.TrimSpace(withoutNulls)
}
