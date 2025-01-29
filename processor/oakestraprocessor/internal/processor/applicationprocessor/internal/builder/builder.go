package builder

import (
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type MetricBuilder struct {
	metrics pmetric.MetricSlice
	mb      *MetricsBuilder
}

func (mb *MetricBuilder) AddGauge(name, state, description, unit string, value float64) *MetricBuilder {
	metric := mb.metrics.AppendEmpty()
	metric.SetName(name)
	metric.SetDescription(description)
	metric.SetUnit(unit)
	dp := metric.SetEmptyGauge().DataPoints().AppendEmpty()
	dp.SetDoubleValue(value)
	dp.SetTimestamp(mb.mb.timestamp)
	dp.Attributes().PutStr("state", state)
	return mb
}

func (mb *MetricBuilder) AddSum(name, state, description, unit string, value float64) *MetricBuilder {
	metric := mb.metrics.AppendEmpty()
	metric.SetName(name)
	metric.SetDescription(description)
	metric.SetUnit(unit)
	sum := metric.SetEmptySum()
	sum.SetIsMonotonic(true)
	sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
	dp := sum.DataPoints().AppendEmpty()
	dp.SetDoubleValue(value)
	dp.SetTimestamp(mb.mb.timestamp)
	dp.Attributes().PutStr("state", state)
	return mb
}
