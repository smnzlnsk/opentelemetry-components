package repository

import (
	"github.com/smnzlnsk/opentelemetry-components/pkg/database"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/domain"
	"go.uber.org/zap"
)

type Repositories struct {
	ContractRepository domain.ContractRepository
	MetricsRepository  domain.MetricsRepository
}

func NewRepositories(client database.Client, logger *zap.Logger) *Repositories {
	return &Repositories{
		ContractRepository: NewContractRepository(client.GetContractStore(), logger),
		MetricsRepository:  NewMetricsRepository(client.GetMetricsStore(), logger),
	}
}
