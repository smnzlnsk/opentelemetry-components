package domain

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/pkg/database"
)

type MetricsRepository interface {
	GetJobMetrics(ctx context.Context, jobName string) (database.HostMetrics, error)
	GetJobMetricsAsMap(ctx context.Context, jobName string) (database.HostMetricsMap, error)
}
