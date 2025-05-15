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
	return &services{
		ContractService: NewContractService(repositories.ContractRepository, logger),
		MetricsService:  NewMetricsService(repositories.MetricsRepository, logger),
	}
}

func (s *services) GetContractService() domain.ContractService {
	return s.ContractService
}

func (s *services) GetMetricsService() domain.MetricsService {
	return s.MetricsService
}
