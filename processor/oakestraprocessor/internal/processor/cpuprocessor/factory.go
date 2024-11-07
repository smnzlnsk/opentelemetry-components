package cpuprocessor

import (
	"context"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
)

const (
	TypeStr = "cpu"
)

var (
	processorType component.Type = component.MustNewType(TypeStr)
)

type Factory struct{}

func (f *Factory) CreateDefaultConfig() internal.Config {
	return Config{}
}

func (f *Factory) CreateMetricsProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg internal.Config,
	logger *zap.Logger,
) (internal.MetricProcessor, error) {
	return newCPUMetricProcessor(ctx, set, cfg, logger)
}
