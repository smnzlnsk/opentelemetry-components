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

type Calculation struct {
	Service string
	Metrics map[string]pmetric.Metric
}

func NewCalculation(filter map[string]bool) *Calculation {
	ca := &Calculation{
		Metrics: make(map[string]pmetric.Metric),
	}
	for s, _ := range filter {
		ca.Metrics[s] = pmetric.NewMetric()
	}
	return ca
}

type MetricProcessor interface {
	IdentifyServices(pmetric.Metrics) []string
	Start(context.Context, component.Host) error
	ProcessMetrics(pmetric.Metrics) error
	Shutdown(context.Context) error
}
