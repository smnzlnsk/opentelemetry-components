package oakestraheuristicengine

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/internal/shared/config"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
)

// NewFactory creates a factory for the oakestraengine processor.
func NewFactory() processor.Factory {
	return processor.NewFactory(
		metadata.Type,
		createDefaultConfig,
		processor.WithMetrics(createMetricsProcessor, metadata.MetricsStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		HTTPServer: config.HTTPServerConfig{
			Enabled: true,
			Port:    8080,
			Host:    "0.0.0.0",
		},
		MongoDB: config.MongoDBConfig{
			Host:     "localhost",
			Port:     27017,
			User:     "",
			Password: "",
		},
	}
}

func createMetricsProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (processor.Metrics, error) {
	config := cfg.(*Config)
	return newProcessor(config, set, nextConsumer)
}
