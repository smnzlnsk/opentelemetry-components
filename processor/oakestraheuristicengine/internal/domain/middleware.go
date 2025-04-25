package domain

import "go.opentelemetry.io/collector/pdata/pmetric"

type MetricsTransformer interface {
	PMetricToMap(md pmetric.Metrics) (map[string]pmetric.Metrics, error)
	ExtractHost(md pmetric.Metrics) string
	TransformToDBHostMetrics(md pmetric.Metrics) (DBHostMetrics, error)
	MergeHostMetrics(existing, new DBHostMetrics) DBHostMetrics
	TransformDBHostMetricsToMap(dbHostMetrics DBHostMetrics) (MapHostMetrics, error)
}
