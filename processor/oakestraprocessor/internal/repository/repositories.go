package repository

import (
	"github.com/smnzlnsk/opentelemetry-components/internal/shared/database"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/domain"
	"go.uber.org/zap"
)

type Repositories struct {
	ContractRepository domain.ContractRepository
	MetricsRepository  domain.MetricsRepository
}

func NewRepositories(client *database.MongoDBClient, logger *zap.Logger) *Repositories {
	return &Repositories{
		ContractRepository: NewContractRepository(client.GetDatabase().Collection("contracts"), logger),
		MetricsRepository:  NewMetricsRepository(client.GetDatabase().Collection("metrics"), logger),
	}
}
