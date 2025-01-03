package oakestraprocessor // import github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor

import (
	"context"
	"fmt"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
	pb "github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/proto"
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
	grpcServer *GRPCServer
}

func newMultiProcessor(ctx context.Context, set processor.Settings, cfg *Config, next consumer.Metrics) *MultiProcessor {
	p, err := createProcessors(ctx, set, cfg, processorFactories)
	if err != nil {
		set.Logger.Error(err.Error())
		return nil
	}

	proc := &MultiProcessor{
		processors: p,
		next:       next,
		logger:     set.Logger,
	}

	// Initialize gRPC server
	proc.grpcServer = NewGRPCServer(proc, cfg.GRPCPort)

	return proc
}

func createProcessors(
	ctx context.Context,
	set processor.Settings,
	config *Config,
	factories map[string]internal.ProcessorFactory,
) ([]internal.MetricProcessor, error) {

	processors := make([]internal.MetricProcessor, 0, len(config.Processors))

	for key, cfg := range config.Processors {
		metricsProcessor, err := createProcessor(ctx, set, cfg, key, factories)
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
	key string,
	factories map[string]internal.ProcessorFactory,
) (internal.MetricProcessor, error) {
	factory := factories[key]
	if factory == nil {
		return nil, fmt.Errorf("unknown processor: %s", key)
	}
	p, err := factory.CreateMetricsProcessor(ctx, set, cfg)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *MultiProcessor) Start(ctx context.Context, host component.Host) error {
	p.logger.Info("Starting Oakestra Processor")
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	p.cancel = cancel

	// Start gRPC server
	if err := p.grpcServer.Start(); err != nil {
		return fmt.Errorf("failed to start gRPC server: %w", err)
	}

	for _, subp := range p.processors {
		if err := subp.Start(ctx, host); err != nil {
			return err
		}
	}

	return nil
}

func (p *MultiProcessor) Shutdown(ctx context.Context) error {
	p.logger.Info("Shutting down Oakestra Processor")

	// Stop gRPC server
	if p.grpcServer != nil {
		p.grpcServer.Stop()
	}

	for _, subp := range p.processors {
		if err := subp.Shutdown(ctx); err != nil {
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

func (p *MultiProcessor) registerService(serviceName string, instanceNumber int32, resource *pb.ResourceInfo) error {
	for _, subp := range p.processors {
		err := subp.RegisterService(serviceName, instanceNumber, resource)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiProcessor) deleteService(serviceName string, instanceNumber int32) error {
	for _, subp := range p.processors {
		err := subp.DeleteService(serviceName, instanceNumber)
		if err != nil {
			return err
		}
	}
	return nil
}
