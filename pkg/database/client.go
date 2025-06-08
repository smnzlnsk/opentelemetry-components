package database

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/pkg/calculation"
	"github.com/smnzlnsk/opentelemetry-components/pkg/contract"
)

// Client is the high-level interface that abstracts away specific database implementations
type Client interface {
	// Connect establishes a connection to the database
	Connect(ctx context.Context) error

	// Close closes the database connection
	Close(ctx context.Context) error

	// Health checks if the database connection is healthy
	Health(ctx context.Context) error

	// GetMetricsStore returns a metrics-specific store interface
	GetMetricsStore() MetricsStore

	// GetContractStore returns a contract-specific store interface
	GetContractStore() ContractStore
}

// MetricsStore provides database operations specifically for metrics
type MetricsStore interface {
	// SaveMetrics saves host metrics to the database
	SaveMetrics(ctx context.Context, metrics HostMetrics) error

	// GetJobMetrics retrieves metrics for a specific job
	GetJobMetrics(ctx context.Context, jobName string) (HostMetrics, error)

	// GetJobMetricsAsMap retrieves metrics for a specific job as a map
	GetJobMetricsAsMap(ctx context.Context, jobName string) (HostMetricsMap, error)

	// DeleteJobMetrics deletes metrics for a specific job
	DeleteJobMetrics(ctx context.Context, jobName string) error

	// DeleteExpiredMetrics deletes metrics older than the specified duration
	DeleteExpiredMetrics(ctx context.Context, maxAge int64) error
}

// ContractStore provides database operations specifically for contracts
type ContractStore interface {
	// Create creates a new contract for a service
	Create(ctx context.Context, contract calculation.Contract) error

	// Update updates an existing contract
	Update(ctx context.Context, old calculation.Contract, new calculation.Contract) error

	// DeleteFormula deletes a specific formula from a service's contracts
	DeleteFormula(ctx context.Context, contract calculation.Contract) error

	// DeleteContract deletes all contracts for a service
	DeleteContract(ctx context.Context, service string) error

	// GetContractsForProcessor retrieves all contracts for a specific processor
	GetContractsForProcessor(ctx context.Context, processor string) ([]contract.Document, error)
}
