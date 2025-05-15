package domain

import "go.opentelemetry.io/collector/pdata/pmetric"

type MetricsTransformer interface {
	PMetricToMap(md pmetric.Metrics) (map[string]pmetric.Metrics, error)
	ExtractHost(md pmetric.Metrics) string
	MergeHostMetrics(existing, new DBHostMetrics) DBHostMetrics
	TransformDBHostMetricsToMap(dbHostMetrics DBHostMetrics) (MapHostMetrics, error)
}
