package database

import (
	"context"
)

// Client is the high-level interface that abstracts away specific database implementations
type Client interface {
	// Connect establishes a connection to the database
	Connect(ctx context.Context) error

	// Close closes the database connection
	Close(ctx context.Context) error

	// Health checks if the database connection is healthy
	Health(ctx context.Context) error

	// GetMetricsStore returns a metrics-specific store interface
	GetMetricsStore() MetricsStore
}

// MetricsStore provides database operations specifically for metrics
type MetricsStore interface {
	// SaveMetrics saves host metrics to the database
	SaveMetrics(ctx context.Context, metrics HostMetrics) error

	// GetJobMetrics retrieves metrics for a specific job
	GetJobMetrics(ctx context.Context, jobName string) (HostMetrics, error)

	// GetJobMetricsAsMap retrieves metrics for a specific job as a map
	GetJobMetricsAsMap(ctx context.Context, jobName string) (HostMetricsMap, error)

	// DeleteJobMetrics deletes metrics for a specific job
	DeleteJobMetrics(ctx context.Context, jobName string) error

	// DeleteExpiredMetrics deletes metrics older than the specified duration
	DeleteExpiredMetrics(ctx context.Context, maxAge int64) error
}
