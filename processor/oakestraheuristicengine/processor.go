package oakestraheuristicengine

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
)

type oakestraheuristicengineProcessor struct {
	config       *Config
	nextConsumer consumer.Metrics
	logger       *zap.Logger
}

func newProcessor(config *Config, set processor.Settings, next consumer.Metrics) (*oakestraheuristicengineProcessor, error) {
	return &oakestraheuristicengineProcessor{
		config:       config,
		nextConsumer: next,
		logger:       set.Logger,
	}, nil
}

func (p *oakestraheuristicengineProcessor) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	// Implement metrics processing logic here
	return p.nextConsumer.ConsumeMetrics(ctx, md)
}

func (p *oakestraheuristicengineProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

func (p *oakestraheuristicengineProcessor) Start(_ context.Context, _ component.Host) error {
	return nil
}

func (p *oakestraheuristicengineProcessor) Shutdown(_ context.Context) error {
	return nil
}
