package service

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/repository"
	"go.uber.org/zap"
)

type services struct {
	ContractService domain.ContractService
	MetricsService  domain.MetricsService
}

func NewServices(repositories *repository.Repositories, logger *zap.Logger) domain.Services {
	// Contract repository should always be available (contracts are always persisted)
	// Only metrics repository is conditional based on PersistentMetrics flag
	var metricsRepository domain.MetricsRepository
	if repositories != nil {
		metricsRepository = repositories.MetricsRepository
	}

	return &services{
		ContractService: NewContractService(repositories.ContractRepository, logger),
		MetricsService:  NewMetricsService(metricsRepository, logger),
	}
}

func (s *services) GetContractService() domain.ContractService {
	return s.ContractService
}

func (s *services) GetMetricsService() domain.MetricsService {
	return s.MetricsService
}
