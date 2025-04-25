package repository

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/persistence/mongodb"
	"go.uber.org/zap"
)

type Repositories struct {
	MetricsRepository domain.MetricsRepository
}

func NewRepositories(client *mongodb.Client, logger *zap.Logger) *Repositories {
	return &Repositories{
		MetricsRepository: NewMetricsRepository(client.GetDatabase().Collection("metrics"), logger),
	}
}
