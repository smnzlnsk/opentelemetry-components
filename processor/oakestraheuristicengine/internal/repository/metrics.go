package repository

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/pkg/database"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/transformers"
	"go.uber.org/zap"
)

// metricsRepository implements domain.MetricsRepository
// It handles storing OpenTelemetry metrics using the abstracted MetricsStore
type metricsRepository struct {
	store       database.MetricsStore
	logger      *zap.Logger
	transformer domain.MetricsTransformer
}

// NewMetricsRepository creates a new metrics repository
func NewMetricsRepository(store database.MetricsStore, logger *zap.Logger) domain.MetricsRepository {
	return &metricsRepository{
		store:       store,
		logger:      logger,
		transformer: transformers.NewMetricsTransformer(logger),
	}
}

// GetJobMetrics gets the metrics for a job
func (r *metricsRepository) GetJobMetrics(ctx context.Context, jobName string) (database.HostMetrics, error) {
	return r.store.GetJobMetrics(ctx, jobName)
}

func (r *metricsRepository) GetJobMetricsAsMap(ctx context.Context, jobName string) (database.HostMetricsMap, error) {
	return r.store.GetJobMetricsAsMap(ctx, jobName)
}
