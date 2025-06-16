package domain

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/pkg/database"
)

type MetricsService interface {
	GetJobMetrics(ctx context.Context, jobName string) (database.HostMetrics, error)
	GetJobMetricsAsMap(ctx context.Context, jobName string) (database.HostMetricsMap, error)
	GetJobMetricsBatch(ctx context.Context, jobNames []string) (map[string]database.HostMetrics, error)
	GetJobMetricsAsMapBatch(ctx context.Context, jobNames []string) (map[string]database.HostMetricsMap, error)
	EnsureIndexes(ctx context.Context) error
}

type Services interface {
	GetMetricsService() MetricsService
}
