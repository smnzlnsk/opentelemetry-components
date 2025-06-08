package repository

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/pkg/calculation"
	"github.com/smnzlnsk/opentelemetry-components/pkg/contract"
	"github.com/smnzlnsk/opentelemetry-components/pkg/database"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/domain"
	"go.uber.org/zap"
)

type contractRepository struct {
	contractStore database.ContractStore
	logger        *zap.Logger
}

func NewContractRepository(contractStore database.ContractStore, logger *zap.Logger) domain.ContractRepository {
	return &contractRepository{
		contractStore: contractStore,
		logger:        logger,
	}
}

func (r *contractRepository) Create(ctx context.Context, ct calculation.Contract) error {
	return r.contractStore.Create(ctx, ct)
}

func (r *contractRepository) Update(ctx context.Context, old calculation.Contract, new calculation.Contract) error {
	return r.contractStore.Update(ctx, old, new)
}

func (r *contractRepository) DeleteFormula(ctx context.Context, ct calculation.Contract) error {
	return r.contractStore.DeleteFormula(ctx, ct)
}

func (r *contractRepository) DeleteContract(ctx context.Context, service string) error {
	return r.contractStore.DeleteContract(ctx, service)
}

func (r *contractRepository) GetContractsForProcessor(ctx context.Context, processor string) ([]contract.Document, error) {
	return r.contractStore.GetContractsForProcessor(ctx, processor)
}
