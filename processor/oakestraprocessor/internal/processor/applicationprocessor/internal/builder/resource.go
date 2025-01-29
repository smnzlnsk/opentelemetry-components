package builder

import (
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type ResourceBuilder struct {
	resource pmetric.ResourceMetrics
	mb       *MetricsBuilder
}

func (rb *ResourceBuilder) SetServiceName(name string) *ResourceBuilder {
	rb.resource.Resource().Attributes().PutStr("service.name", name)
	return rb
}

func (rb *ResourceBuilder) NewMetricBuilder() *MetricBuilder {
	scope := rb.resource.ScopeMetrics().AppendEmpty()
	scope.Scope().SetName("github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/processor/applicationprocessor")
	scope.Scope().SetVersion("0.0.0")
	return &MetricBuilder{
		metrics: scope.Metrics(),
		mb:      rb.mb,
	}
}
