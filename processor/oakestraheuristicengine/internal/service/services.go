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
	return &services{
		MetricsService: NewMetricsService(repositories.MetricsRepository, logger),
	}
}

func (s *services) GetMetricsService() domain.MetricsService {
	return s.MetricsService
}
