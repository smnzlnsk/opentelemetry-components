package memoryprocessor

import (
	"context"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/processor/memoryprocessor/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/processor"
)

const (
	TypeStr = "memory"
)

var (
	processorType component.Type = component.MustNewType(TypeStr)
)

type Factory struct{}

func (f *Factory) CreateDefaultConfig() internal.Config {
	return &Config{
		MetricsBuilderConfig: metadata.DefaultMetricsBuilderConfig(),
	}
}

func (f *Factory) CreateMetricsProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg internal.Config,
) (internal.MetricProcessor, error) {
	return newMemoryMetricProcessor(ctx, set, cfg)
}
