package domain

import (
	"github.com/smnzlnsk/opentelemetry-components/internal/shared/database"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type MetricsTransformer interface {
	PMetricToMap(md pmetric.Metrics) (map[string]pmetric.Metrics, error)
	ExtractHost(md pmetric.Metrics) string
	MergeHostMetrics(existing, new database.HostMetrics) database.HostMetrics
	TransformDBHostMetricsToMap(dbHostMetrics database.HostMetrics) (database.MapHostMetrics, error)
}
