package repository

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/pkg/database"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/domain"
	"go.uber.org/zap"
)

// metricsRepository implements domain.MetricsRepository
// It handles storing OpenTelemetry metrics using the abstracted MetricsStore
type metricsRepository struct {
	store  database.MetricsStore
	logger *zap.Logger
}

// NewMetricsRepository creates a new metrics repository
func NewMetricsRepository(store database.MetricsStore, logger *zap.Logger) domain.MetricsRepository {
	return &metricsRepository{
		store:  store,
		logger: logger,
	}
}

// SaveMetrics saves OpenTelemetry metrics using the abstracted store
func (r *metricsRepository) SaveMetrics(ctx context.Context, hostMetrics database.HostMetrics) error {
	return r.store.SaveMetrics(ctx, hostMetrics)
}

// mergeHostMetrics merges new metrics into existing host metrics
func (r *metricsRepository) mergeHostMetrics(existing, new database.HostMetrics) database.HostMetrics {
	result := existing

	// Helper function to add new datapoints to existing metrics
	mergeMetricDatapoints := func(existing []database.MetricDatapoints, new []database.MetricDatapoints) []database.MetricDatapoints {
		// Create a map for quick lookup by identifier
		metricMap := make(map[string]int)

		for i, dp := range existing {
			key := dp.Identifier.Name + ":" + dp.Identifier.State
			metricMap[key] = i
		}

		// Add new datapoints
		for _, newDP := range new {
			key := newDP.Identifier.Name + ":" + newDP.Identifier.State

			if idx, found := metricMap[key]; found {
				// Append datapoints to existing entry
				existing[idx].Datapoints = append(existing[idx].Datapoints, newDP.Datapoints...)

				// Keep only the latest 5 datapoints to limit data amount
				if len(existing[idx].Datapoints) > 5 {
					existing[idx].Datapoints = existing[idx].Datapoints[len(existing[idx].Datapoints)-5:]
				}
			} else {
				// For new metrics, also ensure we don't exceed 5 datapoints
				if len(newDP.Datapoints) > 5 {
					newDP.Datapoints = newDP.Datapoints[len(newDP.Datapoints)-5:]
				}
				// Add new metric
				existing = append(existing, newDP)
			}
		}

		return existing
	}

	// Merge system metrics
	result.SystemMetrics = mergeMetricDatapoints(result.SystemMetrics, new.SystemMetrics)

	// Merge service metrics
	for _, newService := range new.ServiceInstanceMetrics {
		var found bool

		// Find matching service
		for i, existingService := range result.ServiceInstanceMetrics {
			if existingService.JobName == newService.JobName &&
				existingService.InstanceNumber == newService.InstanceNumber {
				// Merge metrics for this service
				result.ServiceInstanceMetrics[i].Metrics = mergeMetricDatapoints(
					result.ServiceInstanceMetrics[i].Metrics,
					newService.Metrics,
				)
				found = true
				break
			}
		}

		if !found {
			// Add new service
			result.ServiceInstanceMetrics = append(result.ServiceInstanceMetrics, newService)
		}
	}

	return result
}

// GetJobMetrics gets the metrics for a job
func (r *metricsRepository) GetJobMetrics(ctx context.Context, jobName string) (database.HostMetrics, error) {
	return r.store.GetJobMetrics(ctx, jobName)
}
