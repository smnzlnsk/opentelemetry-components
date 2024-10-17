package httpexporter // import github.com/smnzlnsk/opentelemetry-components/exporter/httpexporter

import (
	"context"
	"github.com/smnzlnsk/opentelemetry-components/exporter/httpexporter/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

const (
	defaultIP   = "0.0.0.0"
	defaultPort = 2254
)

// NewFactory creates a factory for httpexporter
func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		metadata.Type,
		createDefaultConfig,
		exporter.WithMetrics(createMetricsExporter, metadata.MetricsStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		EndpointConfig{
			IP:   defaultIP,
			Port: defaultPort,
		},
	}
}

func createMetricsExporter(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Metrics, error) {
	config := cfg.(*Config)
	logger := set.Logger
	he, err := newHTTPExporter(config, logger)
	if err != nil {
		return nil, err
	}

	return exporterhelper.NewMetricsExporter(
		ctx,
		set,
		cfg,
		he.pushMetrics,
		exporterhelper.WithStart(he.start),
		exporterhelper.WithShutdown(he.shutdown),
	)
}
