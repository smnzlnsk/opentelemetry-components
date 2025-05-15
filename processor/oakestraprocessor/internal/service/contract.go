package service

import (
	"context"
	"errors"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/domain"
	"go.uber.org/zap"
)

type contractService struct {
	contractRepository domain.ContractRepository
	logger             *zap.Logger
}

func NewContractService(contractRepository domain.ContractRepository, logger *zap.Logger) domain.ContractService {
	return &contractService{
		contractRepository: contractRepository,
		logger:             logger,
	}
}

func (s *contractService) Create(ctx context.Context, contract domain.CalculationContract) error {
	// Validate input
	if contract.Service == "" {
		return errors.New("service name cannot be empty")
	}
	if contract.Formula == "" {
		return errors.New("formula cannot be empty")
	}

	return s.contractRepository.Create(ctx, contract)
}

func (s *contractService) CreateMany(ctx context.Context, contracts []domain.CalculationContract) error {
	for _, contract := range contracts {
		if err := s.Create(ctx, contract); err != nil {
			return err
		}
	}
	return nil
}

func (s *contractService) Update(ctx context.Context, old domain.CalculationContract, new domain.CalculationContract) error {
	// Validate input
	if old.Service == "" || new.Service == "" {
		return errors.New("service name cannot be empty")
	}
	if old.Formula == "" || new.Formula == "" {
		return errors.New("formula cannot be empty")
	}

	return s.contractRepository.Update(ctx, old, new)
}

func (s *contractService) DeleteFormula(ctx context.Context, contract domain.CalculationContract) error {
	// Validate input
	if contract.Service == "" {
		return errors.New("service name cannot be empty")
	}
	if contract.Formula == "" {
		return errors.New("formula cannot be empty")
	}

	return s.contractRepository.DeleteFormula(ctx, contract)
}

func (s *contractService) DeleteContract(ctx context.Context, service string) error {
	// Validate input
	if service == "" {
		return errors.New("service name cannot be empty")
	}

	return s.contractRepository.DeleteContract(ctx, service)
}

func (s *contractService) GetContractsForProcessor(ctx context.Context, processor string) ([]domain.ContractDocument, error) {
	return s.contractRepository.GetContractsForProcessor(ctx, processor)
}
