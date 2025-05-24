package domain

import (
	"context"
)

type MetricsRepository interface {
	SaveMetrics(ctx context.Context, dbHostMetrics DBHostMetrics) error
	GetJobMetrics(ctx context.Context, jobName string) (DBHostMetrics, error)
}

type ContractRepository interface {
	Create(ctx context.Context, contract CalculationContract) error
	Update(ctx context.Context, old CalculationContract, new CalculationContract) error
	DeleteFormula(ctx context.Context, contract CalculationContract) error
	DeleteContract(ctx context.Context, service string) error
	GetContractsForProcessor(ctx context.Context, processor string) ([]ContractDocument, error)
}
