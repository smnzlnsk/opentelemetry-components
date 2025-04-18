package service

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

// monitoringService implement interfaces.MonitoringService
type monitoringService struct {
	repository domain.MonitoringRepository
	logger     *zap.Logger
}

func NewMonitoringService(repository domain.MonitoringRepository, logger *zap.Logger) domain.MonitoringService {
	return &monitoringService{
		repository: repository,
		logger:     logger,
	}
}

func (s *monitoringService) SaveMetrics(ctx context.Context, md pmetric.Metrics) error {
	return s.repository.SaveMetrics(ctx, md)
}

func (s *monitoringService) GetHostInstanceMetrics(ctx context.Context, host string, serviceInstance string) (domain.DBHostMetrics, error) {
	return s.repository.GetHostInstanceMetrics(ctx, host, serviceInstance)
}
