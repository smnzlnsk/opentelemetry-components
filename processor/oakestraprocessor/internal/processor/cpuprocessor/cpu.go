package cpuprocessor

import (
	"context"
	"fmt"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
)

type CPUMetricProcessor struct {
	internal.BaseProcessor
	logger *zap.Logger
	cancel context.CancelFunc
}

var _ internal.MetricProcessor = (*CPUMetricProcessor)(nil)

func (c *CPUMetricProcessor) ProcessMetrics(metrics pmetric.Metrics) error {
	calc := internal.NewCalculation("[container.cpu.time] / [system.cpu.time]", c.BaseProcessor.Filter)
	_, err := c.processMetrics(calc, metrics)
	if err != nil {
		return err
	}
	// c.logger.Info(st)
	//c.logger.Info("calculation", zap.Any("c", calc.AtomicCalculation))
	return nil
}

func (c *CPUMetricProcessor) processMetrics(calc *internal.Calculation, metrics pmetric.Metrics) (string, error) {
	st := ""
	c.ExtractMetricsIntoCalculation(metrics, calc)
	c.logger.Info("calculation done", zap.Any("result", calc.EvaluateFormula()))

	// for debugging
	st = fmt.Sprintf("%s\nService: %s\nCalcMetrics: %v", st, calc.Service, calc.Metrics)
	for s, m := range calc.Metrics {
		switch m.Type() {
		case pmetric.MetricTypeGauge:
			st = fmt.Sprintf("%s\n\tGauge: %s\n\t\t%d", st, s, m.Gauge().DataPoints().Len())
			for num := 0; num < m.Gauge().DataPoints().Len(); num++ {
				switch m.Gauge().DataPoints().At(num).ValueType().String() {
				case "Int":
					st = fmt.Sprintf("%s\n\t\t\tIntValue: %d\t\t%v", st, m.Gauge().DataPoints().At(num).IntValue(), m.Gauge().DataPoints().At(num).Attributes().AsRaw())
				case "Double":
					st = fmt.Sprintf("%s\n\t\t\tDoubleValue: %v\t\t%v", st, m.Gauge().DataPoints().At(num).DoubleValue(), m.Gauge().DataPoints().At(num).Attributes().AsRaw())
				}
			}
		case pmetric.MetricTypeSum:
			st = fmt.Sprintf("%s\n\tSum: %s\n\t\t%d", st, s, m.Sum().DataPoints().Len())
			for num := 0; num < m.Sum().DataPoints().Len(); num++ {
				switch m.Sum().DataPoints().At(num).ValueType() {
				case pmetric.NumberDataPointValueTypeInt:
					st = fmt.Sprintf("%s\n\t\t\tIntValue: %d\t\t%v", st, m.Sum().DataPoints().At(num).IntValue(), m.Sum().DataPoints().At(num).Attributes().AsRaw())
				case pmetric.NumberDataPointValueTypeDouble:
					st = fmt.Sprintf("%s\n\t\t\tDoubleValue: %v\t\t%v\t%s", st, m.Sum().DataPoints().At(num).DoubleValue(), m.Sum().DataPoints().At(num).Attributes().AsRaw(), m.Unit())
				default:
					continue
				}
			}
		default:
			continue
		}
	}
	return st, nil
}

func (c *CPUMetricProcessor) Shutdown(ctx context.Context) error {
	if c.cancel != nil {
		c.cancel()
	}
	c.logger.Info("Stopped CPU Processor")
	return nil
}

func (c *CPUMetricProcessor) Start(ctx context.Context, _ component.Host) error {
	ctx, c.cancel = context.WithCancel(ctx)
	c.logger.Info("Started CPU Processor")
	return nil
}

func newCPUMetricProcessor(
	_ context.Context,
	_ processor.Settings,
	_ internal.Config,
	logger *zap.Logger,
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
		logger: logger,
	}, nil
}
