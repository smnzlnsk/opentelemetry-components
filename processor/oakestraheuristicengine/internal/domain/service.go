package domain

import (
	"context"

	"go.opentelemetry.io/collector/pdata/pmetric"
)

type MetricsService interface {
	SaveMetrics(ctx context.Context, md pmetric.Metrics) error
	GetJobMetrics(ctx context.Context, jobName string) (DBHostMetrics, error)
	GetJobMetricsAsMap(ctx context.Context, jobName string) (MapHostMetrics, error)
}

type Services interface {
	GetMetricsService() MetricsService
}
