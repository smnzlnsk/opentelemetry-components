package repository

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/persistence/mongodb"
	"go.uber.org/zap"
)

type Repositories struct {
	ContractRepository domain.ContractRepository
	MetricsRepository  domain.MetricsRepository
}

func NewRepositories(client *mongodb.Client, logger *zap.Logger) *Repositories {
	return &Repositories{
		ContractRepository: NewContractRepository(client.GetDatabase().Collection("contracts"), logger),
		MetricsRepository:  NewMetricsRepository(client.GetDatabase().Collection("metrics"), logger),
	}
}
