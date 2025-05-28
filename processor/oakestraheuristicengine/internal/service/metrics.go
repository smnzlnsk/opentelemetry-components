package service

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/pkg/database"
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

func (s *metricsService) GetJobMetrics(ctx context.Context, jobName string) (database.HostMetrics, error) {
	return s.repository.GetJobMetrics(ctx, jobName)
}

func (s *metricsService) GetJobMetricsAsMap(ctx context.Context, jobName string) (database.HostMetricsMap, error) {
	return s.repository.GetJobMetricsAsMap(ctx, jobName)
}
