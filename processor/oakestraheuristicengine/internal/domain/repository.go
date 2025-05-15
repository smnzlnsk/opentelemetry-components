package domain

import (
	"context"
)

type MetricsRepository interface {
	GetJobMetrics(ctx context.Context, jobName string) (DBHostMetrics, error)
	GetJobMetricsAsMap(ctx context.Context, jobName string) (MapHostMetrics, error)
}
