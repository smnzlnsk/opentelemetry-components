package cpuprocessor

import (
	"context"
	"fmt"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/calculation"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
)

type CPUMetricProcessor struct {
	filter internal.Filter
	cancel context.CancelFunc
	logger *zap.Logger
}

func (c *CPUMetricProcessor) ProcessMetrics(metrics pmetric.Metrics) error {
	calc := calculation.NewCalculation(c.filter)
	_, err := c.processMetrics(calc, metrics)
	if err != nil {
		return err
	}
	// c.logger.Info(st)
	c.logger.Info("calculation", zap.Any("c", calc.AtomicCalculation))
	return nil
}

func (c *CPUMetricProcessor) processMetrics(calc *calculation.Calculation, metrics pmetric.Metrics) (string, error) {
	st := ""
	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		rm := metrics.ResourceMetrics().At(i)
		rmAttr := rm.Resource().Attributes().AsRaw()
		// TODO: this can be done better using the built-in .Get()
		if internal.Map_contains(rmAttr, "container_id") && internal.Map_contains(rmAttr, "namespace") {
			s, _ := rmAttr["container_id"]
			calc.Service = fmt.Sprintf("%v", s)
		}
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			smetric := rm.ScopeMetrics().At(j)
			for k := 0; k < smetric.Metrics().Len(); k++ {
				mmetric := smetric.Metrics().At(k)
				if active, ok := c.filter.MetricFilter[mmetric.Name()]; active && ok {
					for x := 0; x < mmetric.Sum().DataPoints().Len(); x++ {
						ndp := mmetric.Sum().DataPoints().At(x)
						for state, _ := range c.filter.StateFilter {
							if s, ok := ndp.Attributes().Get("state"); ok {
								if state == s.Str() {
									md := calculation.CreateMetricDatapoint(mmetric, x)
									calc.SetValue(state, mmetric.Name(), md)
								}
							}
						}
					}
					// for debugging
					mmetric.CopyTo(calc.Metrics[mmetric.Name()])
				}
			}
		}
	}
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

func (c *CPUMetricProcessor) IdentifyServices(pmetric.Metrics) []string {
	return []string{}
}

func newCPUMetricProcessor(
	_ context.Context,
	_ processor.Settings,
	_ internal.Config,
	logger *zap.Logger,
) (internal.MetricProcessor, error) {
	metricFilter := map[string]bool{
		"container.cpu.time":     true,
		"system.cpu.time":        true,
		"system.cpu.utilization": false,
	}
	stateFilter := map[string]bool{
		"system": true,
		"user":   true,
	}
	return &CPUMetricProcessor{
		filter: internal.Filter{
			MetricFilter: metricFilter,
			StateFilter:  stateFilter,
		},
		logger: logger,
	}, nil
}
