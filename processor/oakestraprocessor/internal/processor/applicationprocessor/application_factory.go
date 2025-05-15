package applicationprocessor

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/domain"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/processor"
)

const (
	TypeStr = "application"
)

var (
	processorType component.Type = component.MustNewType(TypeStr)
)

type Factory struct{}

func (f *Factory) CreateDefaultConfig() internal.Config {
	return &Config{}
}

func (f *Factory) CreateMetricsProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg internal.Config,
	services domain.Services,
) (internal.MetricProcessor, error) {
	return newApplicationMetricProcessor(ctx, set, cfg, services)
}
