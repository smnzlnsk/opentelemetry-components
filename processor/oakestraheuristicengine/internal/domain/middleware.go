package domain

import "go.opentelemetry.io/collector/pdata/pmetric"

type MetricsTransformer interface {
	PMetricToMap(md pmetric.Metrics) (map[string]pmetric.Metrics, error)
	ExtractHost(md pmetric.Metrics) string
	TransformToDBHostMetrics(md pmetric.Metrics) (DBHostMetrics, error)
	ExtractDatapoint(metric pmetric.Metric) *DBMetricDatapoint
	MergeHostMetrics(existing, new DBHostMetrics) DBHostMetrics
}
