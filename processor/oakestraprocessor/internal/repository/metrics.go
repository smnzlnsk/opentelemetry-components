package repository

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// metricsRepository implements domain.MetricsRepository
// It handles storing OpenTelemetry metrics in MongoDB
type metricsRepository struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

// NewMetricsRepository creates a new metrics repository
func NewMetricsRepository(collection *mongo.Collection, logger *zap.Logger) domain.MetricsRepository {
	return &metricsRepository{
		collection: collection,
		logger:     logger,
	}
}

// SaveMetrics saves OpenTelemetry metrics to MongoDB
// We assume all metrics in the input are from a single host
func (r *metricsRepository) SaveMetrics(ctx context.Context, hostMetrics domain.DBHostMetrics) error {
	// Create filter for upsert
	filter := bson.M{"host": hostMetrics.Host}

	// Create update options with upsert
	opts := options.Update().SetUpsert(true)

	// First try to find the existing document
	var existingHost domain.DBHostMetrics
	err := r.collection.FindOne(ctx, filter).Decode(&existingHost)

	if err != nil && err != mongo.ErrNoDocuments {
		r.logger.Error("Failed to query existing host metrics", zap.Error(err), zap.String("host", hostMetrics.Host))
		return err
	}

	if err == mongo.ErrNoDocuments {
		// Insert new document
		_, err = r.collection.InsertOne(ctx, hostMetrics)
		if err != nil {
			r.logger.Error("Failed to insert host metrics", zap.Error(err), zap.String("host", hostMetrics.Host))
			return err
		}
	} else {
		// Merge metrics with existing
		merged := r.mergeHostMetrics(existingHost, hostMetrics)

		// Update the document
		update := bson.M{"$set": merged}
		_, err = r.collection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			r.logger.Error("Failed to update host metrics", zap.Error(err), zap.String("host", hostMetrics.Host))
			return err
		}
	}

	return nil
}

// mergeHostMetrics merges new metrics into existing host metrics
func (r *metricsRepository) mergeHostMetrics(existing, new domain.DBHostMetrics) domain.DBHostMetrics {
	result := existing

	// Helper function to add new datapoints to existing metrics
	mergeMetricDatapoints := func(existing []domain.DBMetricDatapoints, new []domain.DBMetricDatapoints) []domain.DBMetricDatapoints {
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
func (r *metricsRepository) GetJobMetrics(ctx context.Context, jobName string) (domain.DBHostMetrics, error) {
	// Create a filter to find hosts that have service instances with the specified job name
	filter := bson.M{"service_instance_metrics.job_name": jobName}

	// Query the database
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		r.logger.Error("Failed to find metrics from MongoDB", zap.Error(err), zap.String("job_name", jobName))
		return domain.DBHostMetrics{}, err
	}
	defer cursor.Close(ctx)

	// Combine all results into a single DBHostMetrics
	result := domain.DBHostMetrics{
		Host:                   "",
		SystemMetrics:          []domain.DBMetricDatapoints{},
		ServiceInstanceMetrics: []domain.DBServiceInstanceMetrics{},
	}

	// Process all hosts that match the filter
	for cursor.Next(ctx) {
		var hostMetrics domain.DBHostMetrics
		if err := cursor.Decode(&hostMetrics); err != nil {
			r.logger.Error("Failed to decode host metrics", zap.Error(err), zap.String("job_name", jobName))
			continue
		}

		// If this is the first host, use its name
		if result.Host == "" {
			result.Host = hostMetrics.Host
		}

		// Add system metrics
		result.SystemMetrics = append(result.SystemMetrics, hostMetrics.SystemMetrics...)

		// Add only service instance metrics that match the jobName
		for _, serviceInstance := range hostMetrics.ServiceInstanceMetrics {
			if serviceInstance.JobName == jobName {
				result.ServiceInstanceMetrics = append(result.ServiceInstanceMetrics, serviceInstance)
			}
		}
	}

	if err := cursor.Err(); err != nil {
		r.logger.Error("Cursor error while getting job metrics", zap.Error(err), zap.String("job_name", jobName))
		return domain.DBHostMetrics{}, err
	}

	// If no data was found
	if result.Host == "" {
		r.logger.Warn("No metrics found for job", zap.String("job_name", jobName))
		return domain.DBHostMetrics{}, mongo.ErrNoDocuments
	}

	return result, nil
}
