package oakestraprocessor // import github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor

import (
	"context"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
)

var _ processor.Metrics = (*Processor)(nil)

type Processor struct {
	next   consumer.Metrics
	log    *zap.Logger
	cancel context.CancelFunc
}

func newProcessor(cfg *Config, logger *zap.Logger, next consumer.Metrics) *Processor {

	proc := Processor{
		next: next,
		log:  logger,
	}
	return &proc
}

func (p *Processor) Start(ctx context.Context, _ component.Host) error {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	p.log.Info("starting oakestra processor")
	return nil
}

func (p *Processor) Shutdown(_ context.Context) error {
	p.cancel()
	return nil
}

func (p *Processor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

func (p *Processor) ConsumeMetrics(ctx context.Context, metrics pmetric.Metrics) error {
	p.log.Info("metrics consumed", zap.Any("metrics", metrics))
	return nil
}
