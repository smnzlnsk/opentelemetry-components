package domain

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/pkg/calculation"
	"github.com/smnzlnsk/opentelemetry-components/pkg/contract"
	"github.com/smnzlnsk/opentelemetry-components/pkg/database"
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
