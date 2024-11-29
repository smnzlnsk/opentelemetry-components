package internal

import (
	"context"
	"errors"
	"fmt"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
)

type ProcessorFactory interface {
	CreateDefaultConfig() Config
	CreateMetricsProcessor(
		ctx context.Context,
		settings processor.Settings,
		cfg Config) (MetricProcessor, error)
}

type MetricProcessor interface {
	Start(context.Context, component.Host) error
	ProcessMetrics(pmetric.Metrics) error
	Shutdown(context.Context) error
	ExtractMetricsIntoCalculation(pmetric.Metrics, *Calculation)
}

type BaseProcessor struct {
	Filter Filter
}

func (b *BaseProcessor) Start(_ context.Context, _ component.Host) error {
	return errors.New("implement me")
}
func (b *BaseProcessor) Shutdown(_ context.Context) error {
	return errors.New("implement me")
}
func (b *BaseProcessor) ProcessMetrics(_ pmetric.Metrics) error {
	return errors.New("implement me")
}

func (b *BaseProcessor) ExtractMetricsIntoCalculation(metrics pmetric.Metrics, calc *Calculation) {
	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		rm := metrics.ResourceMetrics().At(i)
		rmAttr := rm.Resource().Attributes().AsRaw()
		// TODO: this can be done better using the built-in .Get()
		if mapContains(rmAttr, "container_id") && mapContains(rmAttr, "namespace") {
			s, _ := rmAttr["container_id"]
			calc.Service = fmt.Sprintf("%v", s)
		}
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			smetric := rm.ScopeMetrics().At(j)
			for k := 0; k < smetric.Metrics().Len(); k++ {
				mmetric := smetric.Metrics().At(k)
				// guard clause, check that metric is set in filter and active
				if active, ok := b.Filter.MetricFilter[mmetric.Name()]; !ok && active {
					continue
				}
				for x := 0; x < mmetric.Sum().DataPoints().Len(); x++ {
					ndp := mmetric.Sum().DataPoints().At(x)
					// get 'state' attribute, if it exists, otherwise go to next metric
					attr, ok := ndp.Attributes().Get("state")
					if !ok {
						break
					}
					attrStr := attr.Str()
					// see, if found state is supposed to be part of the calculation and set values
					if _, ok := b.Filter.StateFilter[attrStr]; ok {
						md := CreateMetricDatapoint(mmetric, x)
						calc.SetValue(attrStr, mmetric.Name(), md)
					}
				}
			}
		}
	}
}
