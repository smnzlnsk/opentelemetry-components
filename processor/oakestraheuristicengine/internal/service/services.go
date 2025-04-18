package service

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/repository"
	"go.uber.org/zap"
)

type Services struct {
	MonitoringService domain.MonitoringService
}

func NewServices(repositories *repository.Repositories, logger *zap.Logger) *Services {
	return &Services{
		MonitoringService: NewMonitoringService(repositories.MonitoringRepository, logger),
	}
}
