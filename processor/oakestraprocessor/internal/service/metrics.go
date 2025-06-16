package service

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/pkg/database"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/domain"
	"go.uber.org/zap"
)

// monitoringService implement interfaces.MonitoringService
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

func (s *metricsService) SaveMetrics(ctx context.Context, dbHostMetrics database.HostMetrics) error {
	return s.repository.SaveMetrics(ctx, dbHostMetrics)
}

// SaveMetricsBatch saves multiple host metrics in a single batch operation
func (s *metricsService) SaveMetricsBatch(ctx context.Context, hostMetricsList []database.HostMetrics) error {
	// Check if repository supports batch operations
	if batchRepo, ok := s.repository.(interface {
		SaveMetricsBatch(ctx context.Context, hostMetricsList []database.HostMetrics) error
	}); ok {
		return batchRepo.SaveMetricsBatch(ctx, hostMetricsList)
	}

	// Fallback to individual saves
	for _, metrics := range hostMetricsList {
		if err := s.repository.SaveMetrics(ctx, metrics); err != nil {
			s.logger.Error("Failed to save metrics in batch fallback", zap.Error(err))
			return err
		}
	}

	return nil
}

func (s *metricsService) GetJobMetrics(ctx context.Context, jobName string) (database.HostMetrics, error) {
	return s.repository.GetJobMetrics(ctx, jobName)
}

// GetJobMetricsBatch gets metrics for multiple jobs in a single batch operation
func (s *metricsService) GetJobMetricsBatch(ctx context.Context, jobNames []string) (map[string]database.HostMetrics, error) {
	// Check if repository supports batch operations
	if batchRepo, ok := s.repository.(interface {
		GetJobMetricsBatch(ctx context.Context, jobNames []string) (map[string]database.HostMetrics, error)
	}); ok {
		return batchRepo.GetJobMetricsBatch(ctx, jobNames)
	}

	// Fallback to individual gets
	results := make(map[string]database.HostMetrics)
	for _, jobName := range jobNames {
		metrics, err := s.repository.GetJobMetrics(ctx, jobName)
		if err != nil {
			s.logger.Error("Failed to get job metrics in batch fallback", zap.Error(err), zap.String("job_name", jobName))
			return nil, err
		}
		results[jobName] = metrics
	}

	return results, nil
}

// EnsureIndexes ensures database indexes are created for optimal performance
func (s *metricsService) EnsureIndexes(ctx context.Context) error {
	if indexRepo, ok := s.repository.(interface {
		EnsureIndexes(ctx context.Context) error
	}); ok {
		return indexRepo.EnsureIndexes(ctx)
	}

	s.logger.Warn("Repository does not support index creation")
	return nil
}
