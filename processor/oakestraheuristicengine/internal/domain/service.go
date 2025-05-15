package domain

import (
	"context"
)

type MetricsService interface {
	GetJobMetrics(ctx context.Context, jobName string) (DBHostMetrics, error)
	GetJobMetricsAsMap(ctx context.Context, jobName string) (MapHostMetrics, error)
}

type Services interface {
	GetMetricsService() MetricsService
}
