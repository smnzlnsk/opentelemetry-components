package domain

import (
	"context"

	"go.opentelemetry.io/collector/pdata/pmetric"
)

type MetricsRepository interface {
	SaveMetrics(ctx context.Context, md pmetric.Metrics) error
	GetJobMetrics(ctx context.Context, jobName string) (DBHostMetrics, error)
	GetJobMetricsAsMap(ctx context.Context, jobName string) (MapHostMetrics, error)
}

type ContractRepository interface {
	Create(ctx context.Context, contract CalculationContract) error
	Update(ctx context.Context, old CalculationContract, new CalculationContract) error
	DeleteFormula(ctx context.Context, contract CalculationContract) error
	DeleteContract(ctx context.Context, service string) error
	GetContractsForProcessor(ctx context.Context, processor string) ([]ContractDocument, error)
}
