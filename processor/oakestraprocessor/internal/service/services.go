package service

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/repository"
	"go.uber.org/zap"
)

type Services struct {
	ContractService ContractService
}

func NewServices(repositories *repository.Repositories, logger *zap.Logger) *Services {
	return &Services{
		ContractService: NewContractService(repositories.ContractRepository, logger),
	}
}
