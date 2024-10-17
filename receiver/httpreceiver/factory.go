package httpreceiver // import github.com/smnzlnsk/opentelemetry-components/receiver/httpreceiver

import (
	"context"
	"github.com/smnzlnsk/opentelemetry-components/receiver/httpreceiver/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

const ()

// NewFactory creates a factory for httpreceiver
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		metadata.Type,
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, metadata.MetricsStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createMetricsReceiver(
	ctx context.Context,
	set receiver.Settings,
	cfg component.Config,
	con consumer.Metrics,
) (receiver.Metrics, error) {
	logger := set.Logger
	config := cfg.(*Config)

	hr, err := newHTTPReceiver(config, logger, con)
	if err != nil {
		return nil, err
	}
	return hr, nil
}
