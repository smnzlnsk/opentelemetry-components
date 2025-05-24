package domain

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/internal/shared/calculation"
	"github.com/smnzlnsk/opentelemetry-components/internal/shared/contract"
	"github.com/smnzlnsk/opentelemetry-components/internal/shared/database"
)

type Services interface {
	GetMetricsService() MetricsService
	GetContractService() ContractService
}

type MetricsService interface {
	SaveMetrics(ctx context.Context, dbHostMetrics database.HostMetrics) error
	GetJobMetrics(ctx context.Context, jobName string) (database.HostMetrics, error)
}

type ContractService interface {
	Create(ctx context.Context, contract calculation.Contract) error
	CreateMany(ctx context.Context, contracts []calculation.Contract) error
	Update(ctx context.Context, old calculation.Contract, new calculation.Contract) error
	DeleteFormula(ctx context.Context, contract calculation.Contract) error
	DeleteContract(ctx context.Context, service string) error
	GetContractsForProcessor(ctx context.Context, processor string) ([]contract.Document, error)
}
