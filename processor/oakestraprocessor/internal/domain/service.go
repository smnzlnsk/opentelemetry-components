package domain

import (
	"context"
)

type Services interface {
	GetMetricsService() MetricsService
	GetContractService() ContractService
}

type MetricsService interface {
	SaveMetrics(ctx context.Context, dbHostMetrics DBHostMetrics) error
	GetJobMetrics(ctx context.Context, jobName string) (DBHostMetrics, error)
}

type ContractService interface {
	Create(ctx context.Context, contract CalculationContract) error
	CreateMany(ctx context.Context, contracts []CalculationContract) error
	Update(ctx context.Context, old CalculationContract, new CalculationContract) error
	DeleteFormula(ctx context.Context, contract CalculationContract) error
	DeleteContract(ctx context.Context, service string) error
	GetContractsForProcessor(ctx context.Context, processor string) ([]ContractDocument, error)
}
