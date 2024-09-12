package oakestraprocessor // import github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor

import (
	"context"
	"fmt"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
)

const ()

// NewFactory creates a factory for oakestraprocessor
func NewFactory() processor.Factory {
	return processor.NewFactory(
		metadata.Type,
		createDefaultConfig,
		processor.WithMetrics(createMetricsProcessor, component.StabilityLevelDevelopment),
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createMetricsProcessor(
	_ context.Context,
	set processor.Settings,
	cfg component.Config,
	next consumer.Metrics,
) (processor.Metrics, error) {
	logger := set.Logger
	config, ok := cfg.(*Config)
	if !ok {
		return nil, fmt.Errorf("configuration could not be parsed")
	}

	return newProcessor(config, logger, next), nil
}
