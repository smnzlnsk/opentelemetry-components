package repository

import (
	"github.com/smnzlnsk/opentelemetry-components/pkg/database"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"go.uber.org/zap"
)

type Repositories struct {
	MetricsRepository domain.MetricsRepository
}

func NewRepositories(client database.Client, logger *zap.Logger) *Repositories {
	return &Repositories{
		MetricsRepository: NewMetricsRepository(client.GetMetricsStore(), logger),
	}
}
