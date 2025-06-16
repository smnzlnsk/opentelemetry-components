package domain

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/pkg/calculation"
	"github.com/smnzlnsk/opentelemetry-components/pkg/contract"
	"github.com/smnzlnsk/opentelemetry-components/pkg/database"
)

type MetricsRepository interface {
	SaveMetrics(ctx context.Context, dbHostMetrics database.HostMetrics) error
	GetJobMetrics(ctx context.Context, jobName string) (database.HostMetrics, error)
	GetJobMetricsBatch(ctx context.Context, jobNames []string) (map[string]database.HostMetrics, error)
	SaveMetricsBatch(ctx context.Context, hostMetricsList []database.HostMetrics) error
	EnsureIndexes(ctx context.Context) error
}

type ContractRepository interface {
	Create(ctx context.Context, contract calculation.Contract) error
	Update(ctx context.Context, old calculation.Contract, new calculation.Contract) error
	DeleteFormula(ctx context.Context, contract calculation.Contract) error
	DeleteContract(ctx context.Context, service string) error
	GetContractsForProcessor(ctx context.Context, processor string) ([]contract.Document, error)
}
