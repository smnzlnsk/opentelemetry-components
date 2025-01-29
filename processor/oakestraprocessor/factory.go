package oakestraprocessor // import github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor

import (
	"context"
	"fmt"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/metadata"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/processor/applicationprocessor"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/processor/cpuprocessor"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/processor/memoryprocessor"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
)

var (
	processorFactories = map[string]internal.ProcessorFactory{
		cpuprocessor.TypeStr:         &cpuprocessor.Factory{},
		memoryprocessor.TypeStr:      &memoryprocessor.Factory{},
		applicationprocessor.TypeStr: &applicationprocessor.Factory{},
	}
)

// NewFactory creates a factory for oakestraprocessor
func NewFactory() processor.Factory {
	return processor.NewFactory(
		metadata.Type,
		createDefaultConfig,
		processor.WithMetrics(createMetricsProcessor, metadata.MetricsStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createMetricsProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	next consumer.Metrics,
) (processor.Metrics, error) {
	config, ok := cfg.(*Config)
	if !ok {
		return nil, fmt.Errorf("configuration could not be parsed")
	}

	return newMultiProcessor(ctx, set, config, next), nil
}

func getProcessorFactory(key string) (internal.ProcessorFactory, bool) {
	if factory, ok := processorFactories[key]; ok {
		return factory, true
	}
	return nil, false
}
