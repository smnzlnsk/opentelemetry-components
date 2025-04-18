package service

import (
	"context"
	"errors"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/repository"
	"go.uber.org/zap"
)

type ContractService interface {
	Create(ctx context.Context, contract domain.CalculationContract) error
	CreateMany(ctx context.Context, contracts []domain.CalculationContract) error
	Update(ctx context.Context, old domain.CalculationContract, new domain.CalculationContract) error
	DeleteFormula(ctx context.Context, contract domain.CalculationContract) error
	DeleteContract(ctx context.Context, service string) error
}

type contractService struct {
	contractRepository repository.ContractRepository
	logger             *zap.Logger
}

func NewContractService(contractRepository repository.ContractRepository, logger *zap.Logger) ContractService {
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
