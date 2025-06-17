package service

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/pkg/database"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type metricsService struct {
	repository domain.MetricsRepository
	logger     *zap.Logger
}

func NewMetricsService(repository domain.MetricsRepository, logger *zap.Logger) domain.MetricsService {
	return &metricsService{
		repository: repository, // Can be nil if PersistentMetrics is disabled
		logger:     logger,
	}
}

func (s *metricsService) GetJobMetrics(ctx context.Context, jobName string) (database.HostMetrics, error) {
	// First try to get metrics from the memory database (direct processor communication)
	memoryDB := database.GetGlobalMemoryDatabase()
	if memoryMetrics, exists := memoryDB.GetMetrics(jobName); exists {
		// Convert HostMetricsMap back to HostMetrics for compatibility
		// This is a temporary conversion until we fully migrate to HostMetricsMap
		return s.convertMapToHostMetrics(memoryMetrics), nil
	}

	// Fallback to database if not available in memory database and repository is available
	if s.repository != nil {
		s.logger.Info("Fetching job metrics from database (memory database miss)", zap.String("job_name", jobName))
		return s.repository.GetJobMetrics(ctx, jobName)
	}

	// No persistent metrics available
	s.logger.Warn("No metrics found in memory database and persistent metrics disabled",
		zap.String("job_name", jobName))
	return database.HostMetrics{}, mongo.ErrNoDocuments
}

func (s *metricsService) GetJobMetricsAsMap(ctx context.Context, jobName string) (database.HostMetricsMap, error) {
	// First try to get metrics from the memory database (direct processor communication)
	memoryDB := database.GetGlobalMemoryDatabase()
	if memoryMetrics, exists := memoryDB.GetMetrics(jobName); exists {
		return memoryMetrics, nil
	}

	// Fallback to database if not available in memory database and repository is available
	if s.repository != nil {
		s.logger.Info("Fetching job metrics from database (memory database miss)", zap.String("job_name", jobName))
		return s.repository.GetJobMetricsAsMap(ctx, jobName)
	}

	// No persistent metrics available
	s.logger.Warn("No metrics found in memory database and persistent metrics disabled",
		zap.String("job_name", jobName))
	return database.HostMetricsMap{}, mongo.ErrNoDocuments
}

// GetJobMetricsBatch gets metrics for multiple jobs in a single batch operation
func (s *metricsService) GetJobMetricsBatch(ctx context.Context, jobNames []string) (map[string]database.HostMetrics, error) {
	if s.repository != nil {
		return s.repository.GetJobMetricsBatch(ctx, jobNames)
	}

	s.logger.Warn("Batch metrics request with persistent metrics disabled",
		zap.Strings("job_names", jobNames))
	return make(map[string]database.HostMetrics), mongo.ErrNoDocuments
}

// GetJobMetricsAsMapBatch gets metrics as map for multiple jobs in a single batch operation
func (s *metricsService) GetJobMetricsAsMapBatch(ctx context.Context, jobNames []string) (map[string]database.HostMetricsMap, error) {
	if s.repository != nil {
		return s.repository.GetJobMetricsAsMapBatch(ctx, jobNames)
	}

	s.logger.Warn("Batch metrics map request with persistent metrics disabled",
		zap.Strings("job_names", jobNames))
	return make(map[string]database.HostMetricsMap), mongo.ErrNoDocuments
}

// EnsureIndexes ensures database indexes are created for optimal performance
func (s *metricsService) EnsureIndexes(ctx context.Context) error {
	if s.repository != nil {
		return s.repository.EnsureIndexes(ctx)
	}

	s.logger.Info("Skipping database index creation (persistent metrics disabled)")
	return nil
}

// convertMapToHostMetrics converts HostMetricsMap back to HostMetrics for backward compatibility
// This is a temporary method until we fully migrate to HostMetricsMap
func (s *metricsService) convertMapToHostMetrics(hostMetricsMap database.HostMetricsMap) database.HostMetrics {
	result := database.HostMetrics{
		Host:                   "",
		SystemMetrics:          []database.MetricDatapoints{},
		ServiceInstanceMetrics: []database.ServiceInstanceMetrics{},
	}

	// This is a simplified conversion - in practice, we'd want to fully migrate to HostMetricsMap
	// For now, just set a placeholder host name
	for hostName := range hostMetricsMap {
		result.Host = hostName
		break
	}

	return result
}
