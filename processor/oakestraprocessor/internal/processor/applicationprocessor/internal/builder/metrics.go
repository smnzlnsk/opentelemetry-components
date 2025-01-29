package builder

import (
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type MetricsBuilder struct {
	metrics   pmetric.Metrics
	timestamp pcommon.Timestamp
}

func NewMetricsBuilder() *MetricsBuilder {
	return &MetricsBuilder{
		metrics:   pmetric.NewMetrics(),
		timestamp: pcommon.Timestamp(time.Now().UnixNano()),
	}
}

func (b *MetricsBuilder) NewResourceBuilder() *ResourceBuilder {
	return &ResourceBuilder{
		resource: b.metrics.ResourceMetrics().AppendEmpty(),
		mb:       b,
	}
}

func (b *MetricsBuilder) Emit() pmetric.Metrics {
	metrics := b.metrics
	b.metrics = pmetric.NewMetrics()
	return metrics
}
