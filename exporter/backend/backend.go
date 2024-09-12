package backendexporter // import github.com/smnzlnsk/opentelemetry-components/exporter/backend

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type backend struct {
	func_table map[string]func(pmetric.Metric)
	config     *Config
	logger     *zap.Logger
	*marshaler
	host   component.Host
	cancel context.CancelFunc
}

func newBackend(cfg *Config, logger *zap.Logger) (*backend, error) {
	backend := &backend{
		config: cfg,
		logger: logger,
	}
	backend.func_table = map[string]func(pmetric.Metric){
		"Gauge":     backend.handleGauge,
		"Sum":       backend.handleSum,
		"Histogram": backend.handleHistogram,
	}
	return backend, nil
}

func (b *backend) start(ctx context.Context, host component.Host) error {
	ctx = context.Background()
	ctx, b.cancel = context.WithCancel(ctx)
	marshaler, err := newMarshaler()
	if err != nil {
		return err
	}
	b.marshaler = marshaler
	b.host = host

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (b *backend) shutdown(ctx context.Context) error {
	if b.cancel != nil {
		b.cancel()
	}
	return nil
}

func (b *backend) pushMetrics(ctx context.Context, md pmetric.Metrics) error {
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		resourceMetrics := md.ResourceMetrics().At(i)
		for j := 0; j < resourceMetrics.ScopeMetrics().Len(); j++ {
			scopeMetrics := resourceMetrics.ScopeMetrics().At(j)
			for k := 0; k < scopeMetrics.Metrics().Len(); k++ {
				metric := scopeMetrics.Metrics().At(k)
				// handle metric through direct call of the lookup_table
				b.func_table[metric.Type().String()](metric)
			}
		}
	}
	b.logger.Debug("received metric data")
	return nil
}

func (b *backend) handleGauge(metric pmetric.Metric) {
	for i := 0; i < metric.Gauge().DataPoints().Len(); i++ {
		dataPoint := metric.Gauge().DataPoints().At(i)
		b.logger.Info(fmt.Sprintf("%s, %s, %f", metric.Type().String(), metric.Name(), dataPoint.DoubleValue()))
	}
}

func (b *backend) handleSum(metric pmetric.Metric) {
	for i := 0; i < metric.Sum().DataPoints().Len(); i++ {
		dataPoint := metric.Sum().DataPoints().At(i)
		machine, _ := dataPoint.Attributes().Get("machine")

		switch dataPoint.ValueType().String() {
		case "Double":
			b.logger.Info(fmt.Sprintf("%s, %s, %s, %f, %s", metric.Type().String(), metric.Name(), dataPoint.ValueType().String(), dataPoint.DoubleValue(), machine.Str()))
		case "Int":
			b.logger.Info(fmt.Sprintf("%s, %s, %s, %d, %s", metric.Type().String(), metric.Name(), dataPoint.ValueType().String(), dataPoint.IntValue(), machine.Str()))
		default:
			b.logger.Debug("unsupported datapoint type")
			return
		}
	}
}

func (b *backend) handleHistogram(metric pmetric.Metric) {
	for i := 0; i < metric.Histogram().DataPoints().Len(); i++ {
		dataPoint := metric.Histogram().DataPoints().At(i)
		// TODO: Implement me!
		b.logger.Info(fmt.Sprintf("%s, %s, %v", metric.Type().String(), metric.Name(), dataPoint.Attributes()))
	}
}
