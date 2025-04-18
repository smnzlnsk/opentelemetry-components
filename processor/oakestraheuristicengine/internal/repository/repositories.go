package repository

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/persistence/mongodb"
	"go.uber.org/zap"
)

type Repositories struct {
	MonitoringRepository domain.MonitoringRepository
}

func NewRepositories(client *mongodb.Client, logger *zap.Logger) *Repositories {
	return &Repositories{
		MonitoringRepository: NewMonitoringRepository(client.GetDatabase().Collection("metrics"), logger),
	}
}
