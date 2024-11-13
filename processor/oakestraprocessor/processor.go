package oakestraprocessor // import github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor

import (
	"context"
	"fmt"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
)

var _ processor.Metrics = (*MultiProcessor)(nil)

type MultiProcessor struct {
	processors []internal.MetricProcessor
	next       consumer.Metrics
	logger     *zap.Logger
	cancel     context.CancelFunc
}

func newMultiProcessor(ctx context.Context, set processor.Settings, cfg *Config, logger *zap.Logger, next consumer.Metrics) *MultiProcessor {
	p, err := createProcessors(ctx, set, cfg, logger, processorFactories)
	if err != nil {
		logger.Error(err.Error())
		return nil
	}

	proc := MultiProcessor{
		processors: p,
		next:       next,
		logger:     logger,
	}
	return &proc
}

func createProcessors(
	ctx context.Context,
	set processor.Settings,
	config *Config,
	logger *zap.Logger,
	factories map[string]internal.ProcessorFactory,
) ([]internal.MetricProcessor, error) {

	processors := make([]internal.MetricProcessor, 0, len(config.Processors))

	for key, cfg := range config.Processors {
		metricsProcessor, err := createProcessor(ctx, set, cfg, logger, key, factories)
		if err != nil {
			return nil, fmt.Errorf("failed to create metrics processor for key: %q: %w", key, err)
		}
		processors = append(processors, metricsProcessor)
	}
	return processors, nil
}

func createProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg internal.Config,
	logger *zap.Logger,
	key string,
	factories map[string]internal.ProcessorFactory,
) (internal.MetricProcessor, error) {
	factory := factories[key]
	if factory == nil {
		return nil, fmt.Errorf("unknown processor: %s", key)
	}
	p, err := factory.CreateMetricsProcessor(ctx, set, cfg, logger)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *MultiProcessor) Start(ctx context.Context, host component.Host) error {
	p.logger.Info("Starting Oakestra Processor")
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	for _, subp := range p.processors {
		err := subp.Start(ctx, host)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *MultiProcessor) Shutdown(ctx context.Context) error {
	p.logger.Info("Shutting down Oakestra Processor")
	for _, subp := range p.processors {
		err := subp.Shutdown(ctx)
		if err != nil {
			return err
		}
	}
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}

func (p *MultiProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

func (p *MultiProcessor) ConsumeMetrics(ctx context.Context, metrics pmetric.Metrics) error {
	p.logger.Info("metrics consumed", zap.Any("metrics", metrics.ResourceMetrics()))
	// TODO: identify service being monitored

	// TODO: compact down metrics into packages of service

	for _, subp := range p.processors {
		err := subp.ProcessMetrics(metrics)
		if err != nil {
			p.logger.Error("error", zap.Error(err))
			return err
		}
	}

	err := p.next.ConsumeMetrics(ctx, metrics)
	if err != nil {
		p.logger.Error("failed to consume metrics", zap.Error(err))
	}
	return nil
}
