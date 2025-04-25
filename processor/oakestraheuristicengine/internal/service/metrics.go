package service

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"go.opentelemetry.io/collector/pdata/pmetric"
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

func (s *metricsService) SaveMetrics(ctx context.Context, md pmetric.Metrics) error {
	return s.repository.SaveMetrics(ctx, md)
}

func (s *metricsService) GetJobMetrics(ctx context.Context, jobName string) (domain.DBHostMetrics, error) {
	return s.repository.GetJobMetrics(ctx, jobName)
}

func (s *metricsService) GetJobMetricsAsMap(ctx context.Context, jobName string) (domain.MapHostMetrics, error) {
	return s.repository.GetJobMetricsAsMap(ctx, jobName)
}
