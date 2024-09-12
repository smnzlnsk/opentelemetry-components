package backendexporter // import github.com/smnzlnsk/opentelemetry-components/exporters/backend

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/exporters/backend/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

// NewFactory creates a factory for mqttexporter-exporter
func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		metadata.Type,
		createDefaultConfig,
		exporter.WithMetrics(createMetricsExporter, component.StabilityLevelDevelopment),
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createMetricsExporter(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Metrics, error) {
	logger := set.Logger
	config := (cfg.(*Config))
	b, err := newBackend(config, logger)
	if err != nil {
		return nil, err
	}
	return exporterhelper.NewMetricsExporter(
		ctx,
		set,
		cfg,
		b.pushMetrics,
		exporterhelper.WithStart(b.start),
		exporterhelper.WithShutdown(b.shutdown),
	)
}
