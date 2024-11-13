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
	filter map[string]bool
	cancel context.CancelFunc
	logger *zap.Logger
}

func (c *CPUMetricProcessor) ProcessMetrics(metrics pmetric.Metrics) error {
	calc := internal.NewCalculation(c.filter)
	st := ""
	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		rm := metrics.ResourceMetrics().At(i)
		rmAttr := rm.Resource().Attributes().AsRaw()
		//c.logger.Info("ResourceMetrics", zap.Any("attributes", rmAttr))
		//s = fmt.Sprintf("%s\nResourceMetrics: %v", s, rmAttr)
		if internal.Map_contains(rmAttr, "container_id") && internal.Map_contains(rmAttr, "namespace") {
			s, _ := rmAttr["container_id"]
			calc.Service = fmt.Sprintf("%v", s)
		}
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			smetric := rm.ScopeMetrics().At(j)
			//c.logger.Info("ScopeMetrics", zap.Any("length", smetric.Metrics().Len()))
			//s = fmt.Sprintf("%s\n\tScopeMetrics: %v", s, smetric.Metrics().Len())
			for k := 0; k < smetric.Metrics().Len(); k++ {
				mmetric := smetric.Metrics().At(k)
				//c.logger.Info("Metric", zap.Any("name", mmetric.Name()), zap.Any("desc", mmetric.Description()))
				//s = fmt.Sprintf("%s\n\t\tMetricName: %v\n\t\tDesc: %v", s, mmetric.Name(), mmetric.Description())
				if _, exists := c.filter[mmetric.Name()]; exists {
					mmetric.CopyTo(calc.Metrics[mmetric.Name()])
				}
			}
		}
	}
	c.logger.Info("calculation", zap.Any("assembled", calc.Metrics), zap.Any("service", calc.Service))
	st = fmt.Sprintf("%s\nService: %s\nCalcMetrics: %v", st, calc.Service, calc.Metrics)
	for s, m := range calc.Metrics {
		switch m.Type().String() {
		case "Summary":
			st = fmt.Sprintf("%s\n\tSummary: %s\n\t\t%d", st, s, m.Summary().DataPoints().Len())
		case "Gauge":
			st = fmt.Sprintf("%s\n\tGauge: %s\n\t\t%d", st, s, m.Gauge().DataPoints().Len())
			for num := 0; num < m.Gauge().DataPoints().Len(); num++ {
				switch m.Gauge().DataPoints().At(num).ValueType().String() {
				case "Int":
					//c.logger.Info("int", zap.Any("v", m.Sum().DataPoints().At(num).IntValue()), zap.Any("state", m.Sum().DataPoints().At(num).Attributes().AsRaw()))
					st = fmt.Sprintf("%s\n\t\t\tIntValue: %d\t\t%v", st, m.Gauge().DataPoints().At(num).IntValue(), m.Gauge().DataPoints().At(num).Attributes().AsRaw())
				case "Double":
					//c.logger.Info("double", zap.Any("v", m.Sum().DataPoints().At(num).DoubleValue()), zap.Any("state", m.Sum().DataPoints().At(num).Attributes().AsRaw()))
					st = fmt.Sprintf("%s\n\t\t\tDoubleValue: %v\t\t%v", st, m.Gauge().DataPoints().At(num).DoubleValue(), m.Gauge().DataPoints().At(num).Attributes().AsRaw())
				case "Empty:":
					//c.logger.Info("Got empty")
				}
			}
		case "Sum":
			st = fmt.Sprintf("%s\n\tSum: %s\n\t\t%d", st, s, m.Sum().DataPoints().Len())
			for num := 0; num < m.Sum().DataPoints().Len(); num++ {
				switch m.Sum().DataPoints().At(num).ValueType().String() {
				case "Int":
					//c.logger.Info("int", zap.Any("v", m.Sum().DataPoints().At(num).IntValue()), zap.Any("state", m.Sum().DataPoints().At(num).Attributes().AsRaw()))
					st = fmt.Sprintf("%s\n\t\t\tIntValue: %d\t\t%v", st, m.Sum().DataPoints().At(num).IntValue(), m.Sum().DataPoints().At(num).Attributes().AsRaw())
				case "Double":
					//c.logger.Info("double", zap.Any("v", m.Sum().DataPoints().At(num).DoubleValue()), zap.Any("state", m.Sum().DataPoints().At(num).Attributes().AsRaw()))
					st = fmt.Sprintf("%s\n\t\t\tDoubleValue: %v\t\t%v", st, m.Sum().DataPoints().At(num).DoubleValue(), m.Sum().DataPoints().At(num).Attributes().AsRaw())
				case "Empty:":
					//c.logger.Info("Got empty")
				}
			}
		case "Histogram":
			//c.logger.Info("Got a histogram", zap.String("metric_name", s), zap.Any("metric", m.Histogram().DataPoints().At(0)))
		}
	}
	c.logger.Info(st)
	return nil
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

func (c *CPUMetricProcessor) IdentifyServices(pmetric.Metrics) []string {
	return []string{}
}

func newCPUMetricProcessor(
	_ context.Context,
	_ processor.Settings,
	_ internal.Config,
	logger *zap.Logger,
) (internal.MetricProcessor, error) {
	filter := map[string]bool{
		"system.cpu.time":                        true,
		"system.cpu.utilization":                 true,
		"container_cpu_usage_usec_microseconds":  true,
		"container_cpu_system_usec_microseconds": true,
		"container_cpu_user_usec_microseconds":   true,
	}
	return &CPUMetricProcessor{
		filter: filter,
		logger: logger,
	}, nil
}
