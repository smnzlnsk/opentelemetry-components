package service

import (
	"context"

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

func (s *metricsService) SaveMetrics(ctx context.Context, dbHostMetrics domain.DBHostMetrics) error {
	return s.repository.SaveMetrics(ctx, dbHostMetrics)
}

func (s *metricsService) GetJobMetrics(ctx context.Context, jobName string) (domain.DBHostMetrics, error) {
	return s.repository.GetJobMetrics(ctx, jobName)
}
