package domain

import (
	"github.com/smnzlnsk/opentelemetry-components/pkg/database"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type MetricsTransformer interface {
	ExtractHost(md pmetric.Metrics) string
	MergeHostMetrics(existing, new database.HostMetrics) database.HostMetrics
	TransformHostMetricsToMap(dbHostMetrics database.HostMetrics) (database.HostMetricsMap, error)
}
