package cpuprocessor

import (
	"context"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
)

type CPUMetricProcessor struct {
	filter []string
	cancel context.CancelFunc
	logger *zap.Logger
}

func (c *CPUMetricProcessor) ProcessMetrics(metrics pmetric.Metrics) error {
	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		rm := metrics.ResourceMetrics().At(i)
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			smetric := rm.ScopeMetrics().At(j)
			for k := 0; k < smetric.Metrics().Len(); k++ {
				mmetric := smetric.Metrics().At(k)
				c.logger.Info(mmetric.Name())
			}
		}
	}
	return nil
}

func (c *CPUMetricProcessor) Shutdown(ctx context.Context) error {
	if c.cancel != nil {
		c.cancel()
	}
	return nil
}

func (c *CPUMetricProcessor) Start(ctx context.Context, _ component.Host) error {
	ctx, c.cancel = context.WithCancel(ctx)
	return nil
}

func (c *CPUMetricProcessor) IdentifyServices(pmetric.Metrics) []string {
	return ""
}

func newCPUMetricProcessor(
	_ context.Context,
	_ processor.Settings,
	_ internal.Config,
	logger *zap.Logger,
) (internal.MetricProcessor, error) {
	return &CPUMetricProcessor{
		filter: []string{
			"process.cpu.time",
			"container_cpu_usage_usec_microseconds",
			"container_cpu_system_usec_microseconds",
			"container_cpu_user_usec_microseconds",
		},
		logger: logger,
	}, nil
}
