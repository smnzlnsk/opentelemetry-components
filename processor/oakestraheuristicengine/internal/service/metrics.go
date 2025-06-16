package service

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/pkg/database"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"go.uber.org/zap"
)

type metricsService struct {
	repository domain.MetricsRepository
	logger     *zap.Logger
}

func NewMetricsService(repository domain.MetricsRepository, logger *zap.Logger) domain.MetricsService {
	return &metricsService{
		repository: repository,
		logger:     logger,
	}
}

func (s *metricsService) GetJobMetrics(ctx context.Context, jobName string) (database.HostMetrics, error) {
	// First try to get metrics from the broker (direct processor communication)
	broker := database.GetGlobalMetricsBroker()
	if brokerMetrics, exists := broker.GetMetrics(jobName); exists {
		s.logger.Debug("Retrieved job metrics from broker", zap.String("job_name", jobName))
		return brokerMetrics, nil
	}

	// Fallback to database if not available in broker
	s.logger.Debug("Fetching job metrics from database (broker miss)", zap.String("job_name", jobName))
	return s.repository.GetJobMetrics(ctx, jobName)
}

func (s *metricsService) GetJobMetricsAsMap(ctx context.Context, jobName string) (database.HostMetricsMap, error) {
	// First try to get metrics from the broker (direct processor communication)
	broker := database.GetGlobalMetricsBroker()
	if brokerMetrics, exists := broker.GetMetricsAsMap(jobName); exists {
		s.logger.Info("Retrieved job metrics from broker",
			zap.String("job_name", jobName),
			zap.Int("hosts", len(brokerMetrics)))
		return brokerMetrics, nil
	}

	// Fallback to database if not available in broker
	s.logger.Info("Fetching job metrics from database (broker miss)", zap.String("job_name", jobName))
	return s.repository.GetJobMetricsAsMap(ctx, jobName)
}

// GetJobMetricsBatch gets metrics for multiple jobs in a single batch operation
func (s *metricsService) GetJobMetricsBatch(ctx context.Context, jobNames []string) (map[string]database.HostMetrics, error) {
	return s.repository.GetJobMetricsBatch(ctx, jobNames)
}

// GetJobMetricsAsMapBatch gets metrics as map for multiple jobs in a single batch operation
func (s *metricsService) GetJobMetricsAsMapBatch(ctx context.Context, jobNames []string) (map[string]database.HostMetricsMap, error) {
	s.logger.Debug("Batch request, fetching from database",
		zap.Strings("job_names", jobNames))
	return s.repository.GetJobMetricsAsMapBatch(ctx, jobNames)
}

// EnsureIndexes ensures database indexes are created for optimal performance
func (s *metricsService) EnsureIndexes(ctx context.Context) error {
	return s.repository.EnsureIndexes(ctx)
}
