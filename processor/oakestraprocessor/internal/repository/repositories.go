package repository

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/persistence/mongodb"
	"go.uber.org/zap"
)

type Repositories struct {
	ContractRepository ContractRepository
}

func NewRepositories(client *mongodb.Client, logger *zap.Logger) *Repositories {
	return &Repositories{
		ContractRepository: NewContractRepository(client.GetDatabase().Collection("contracts"), logger),
	}
}
