package internal

import (
	"context"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
)

type ProcessorFactory interface {
	CreateDefaultConfig() Config
	CreateMetricsProcessor(
		ctx context.Context,
		settings processor.Settings,
		cfg Config,
		logger *zap.Logger) (MetricProcessor, error)
}

// Config is the configuration of a processor
type Config interface {
}

type ProcessorConfig struct {
}

type MetricProcessor interface {
	IdentifyServices(pmetric.Metrics) []string
	Start(context.Context, component.Host) error
	ProcessMetrics(pmetric.Metrics) error
	Shutdown(context.Context) error
}
