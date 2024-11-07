package internal

import "go.opentelemetry.io/collector/pdata/pmetric"

type MetricFilter interface {
	FilterByName(string, pmetric.Metrics) pmetric.Metrics
	GetFilterMetrics() []string
}
