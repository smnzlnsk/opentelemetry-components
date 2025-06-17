package service

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/repository"
	"go.uber.org/zap"
)

type services struct {
	MetricsService domain.MetricsService
}

func NewServices(repositories *repository.Repositories, logger *zap.Logger) *services {
	var metricsRepository domain.MetricsRepository
	if repositories != nil {
		metricsRepository = repositories.MetricsRepository
	}
	// metricsRepository will be nil if repositories is nil (PersistentMetrics disabled)

	return &services{
		MetricsService: NewMetricsService(metricsRepository, logger),
	}
}

func (s *services) GetMetricsService() domain.MetricsService {
	return s.MetricsService
}
