package service

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
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

func (s *metricsService) GetJobMetrics(ctx context.Context, jobName string) (domain.DBHostMetrics, error) {
	return s.repository.GetJobMetrics(ctx, jobName)
}

func (s *metricsService) GetJobMetricsAsMap(ctx context.Context, jobName string) (domain.MapHostMetrics, error) {
	return s.repository.GetJobMetricsAsMap(ctx, jobName)
}
