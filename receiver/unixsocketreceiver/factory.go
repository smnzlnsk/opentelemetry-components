package unixsocketreceiver // import github.com/smnzlnsk/opentelemetry-components/receiver/unixsocketreceiver

import (
	"context"
	"github.com/smnzlnsk/opentelemetry-components/receiver/unixsocketreceiver/internal/metadata"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

const (
	defaultFolder   = "/var/run/unixsocketreceiver"
	defaultInterval = time.Second * 5
)

// NewFactory creates a factory for unixsocketreceiver
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		metadata.Type,
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, metadata.MetricsStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		Folder:   defaultFolder,
		Interval: defaultInterval.String(),
	}
}

func createMetricsReceiver(
	ctx context.Context,
	set receiver.Settings,
	cfg component.Config,
	con consumer.Metrics,
) (receiver.Metrics, error) {
	logger := set.Logger
	config := cfg.(*Config)

	usr, err := newUnixSocketReceiver(config, logger, con)
	if err != nil {
		return nil, err
	}
	return usr, nil
}
