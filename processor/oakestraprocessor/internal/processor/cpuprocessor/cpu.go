package cpuprocessor

import (
	"context"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/processor/cpuprocessor/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
	"time"
)

type CPUMetricProcessor struct {
	internal.BaseProcessor
	config   *Config
	logger   *zap.Logger
	cancel   context.CancelFunc
	settings processor.Settings
	mb       *metadata.MetricsBuilder
}

var _ internal.MetricProcessor = (*CPUMetricProcessor)(nil)

func (c *CPUMetricProcessor) ProcessMetrics(metrics pmetric.Metrics) error {
	calc := internal.NewCalculation("[container.cpu.time] / [system.cpu.time]", c.BaseProcessor.Filter)

	m, err := c.processMetrics(calc, metrics)
	if err != nil {
		return err
	}

	m.ResourceMetrics().MoveAndAppendTo(metrics.ResourceMetrics())
	return nil
}

func (c *CPUMetricProcessor) processMetrics(calc *internal.Calculation, metrics pmetric.Metrics) (pmetric.Metrics, error) {
	c.ExtractMetricsIntoCalculation(metrics, calc)
	for k, v := range calc.EvaluateFormula() {
		if k == "default" {
			continue
		}
		c.logger.Info("Creating metric", zap.Any("state", k), zap.Any("value", v), zap.Any("service", calc.Service))
		c.mb.RecordServiceCPUUtilisationDataPoint(pcommon.NewTimestampFromTime(time.Now()), v.(float64), metadata.MapAttributeState[k])
	}

	// set resources
	rb := c.mb.NewResourceBuilder()
	rb.SetServiceName(calc.Service)
	c.mb.EmitForResource(metadata.WithResource(rb.Emit()))

	return c.mb.Emit(), nil
}

func (c *CPUMetricProcessor) Shutdown(_ context.Context) error {
	if c.cancel != nil {
		c.cancel()
	}
	c.logger.Info("Stopped CPU Processor")
	return nil
}

func (c *CPUMetricProcessor) Start(ctx context.Context, _ component.Host) error {
	ctx, c.cancel = context.WithCancel(ctx)
	// initialize metric builder
	c.mb = metadata.NewMetricsBuilder(c.config.MetricsBuilderConfig, receiver.Settings{TelemetrySettings: c.settings.TelemetrySettings})
	c.logger.Info("Started CPU Processor")
	return nil
}

func newCPUMetricProcessor(
	_ context.Context,
	set processor.Settings,
	cfg internal.Config,
) (internal.MetricProcessor, error) {
	metricFilter := map[string]bool{
		"container.cpu.time":       true,
		"system.cpu.time":          true,
		"system.cpu.logical.count": true,
		"system.cpu.utilization":   false,
	}
	stateFilter := map[string]bool{
		"system": true,
		"user":   true,
	}
	return &CPUMetricProcessor{
		BaseProcessor: internal.BaseProcessor{
			Filter: internal.Filter{
				MetricFilter: metricFilter,
				StateFilter:  stateFilter,
			},
		},
		config:   cfg.(*Config),
		settings: set,
		logger:   set.Logger,
	}, nil
}
