package oakestraprocessor // import github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor

import (
	"context"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
	"sync"
)

var _ processor.Metrics = (*Processor)(nil)

type Processor struct {
	next consumer.Metrics

	log    *zap.Logger
	ctx    context.Context
	cancel context.CancelFunc

	mtx sync.Mutex
}

func newProcessor(cfg *Config, logger *zap.Logger, next consumer.Metrics) *Processor {
	ctx, cancel := context.WithCancel(context.Background())

	proc := Processor{
		next: next,

		log:    logger,
		ctx:    ctx,
		cancel: cancel,
	}
	return &proc
}

func (p *Processor) Start(_ context.Context, _ component.Host) error {
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
	// TODO: implement metrics consumption
	// TODO: extract metrics of interest
	// TODO: apply mutation on metrics of interest
	return nil
}
