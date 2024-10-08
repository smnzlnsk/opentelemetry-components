package unixsocketreceiver // import github.com/smnzlnsk/opentelemetry-components/receiver/unixsocketreceiver
import (
	"context"
	"github.com/smnzlnsk/opentelemetry-components/receiver/unixsocketreceiver/internal/dirwatcher"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
	"net"
	"time"
)

var _ receiver.Metrics = (*unixSocketReceiver)(nil)

type unixSocketReceiver struct {
	config   *Config
	logger   *zap.Logger
	consumer consumer.Metrics
	host     component.Host
	cancel   context.CancelFunc
}

func newUnixSocketReceiver(config *Config, logger *zap.Logger, consumer consumer.Metrics) (*unixSocketReceiver, error) {
	usr := &unixSocketReceiver{
		config:   config,
		logger:   logger,
		consumer: consumer,
	}
	return usr, nil
}

func (usr *unixSocketReceiver) Start(ctx context.Context, host component.Host) error {
	ctx = context.Background()
	ctx, usr.cancel = context.WithCancel(ctx)
	usr.host = host
	usr.logger.Info("Starting Unix Socket Receiver")

	parsedInterval, err := time.ParseDuration(usr.config.Interval)
	if err != nil {
		return err
	}

	dirWatcher := dirwatcher.NewDirectoryWatcher(usr.config.Folder, usr.logger, parsedInterval, usr.connectToUnixSocket)
	dirWatcher.Start()

	return nil
}

func (usr *unixSocketReceiver) Shutdown(_ context.Context) error {
	if usr.cancel != nil {
		usr.cancel()
	}
	usr.logger.Info("Shutdown Unix Socket Receiver")
	return nil
}

func (usr *unixSocketReceiver) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	if usr.consumer == nil {
		usr.logger.Error("no next consumer available for unix socket receiver")
	}

	err := usr.consumer.ConsumeMetrics(ctx, md)
	if err != nil {
		usr.logger.Error("failed to consume metrics", zap.Error(err))
	}
	return nil
}

func (usr *unixSocketReceiver) connectToUnixSocket(socketPath string) {
	// err was already validated in config validation on init
	t, _ := time.ParseDuration(usr.config.Interval)

	for {
		// connect to socket
		conn, err := net.Dial("unix", socketPath)
		if err != nil {
			usr.logger.Error("failed to connect to unix socket",
				zap.String("socket", socketPath),
				zap.Error(err))
			conn.Close()
			return
		}

		usr.logger.Info("connected to new unix socket", zap.String("socket", socketPath))

		// message
		message := "ping"
		_, err = conn.Write([]byte(message))
		if err != nil {
			usr.logger.Error("failed to write to unix socket",
				zap.String("socket", socketPath),
				zap.Error(err))
			conn.Close()
			return
		}

		// response
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			usr.logger.Error("failed to read from unix socket",
				zap.String("socket", socketPath),
				zap.Error(err))
			return
			conn.Close()
		}
		usr.logger.Info("received response from unix socket",
			zap.String("socket", socketPath),
			zap.String("message", string(buffer[:n])))

		time.Sleep(t)
	}
}
