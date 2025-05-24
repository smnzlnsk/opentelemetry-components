package repository

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/internal/shared/database"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/transformers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// metricsRepository implements domain.MetricsRepository
// It handles storing OpenTelemetry metrics in MongoDB
type metricsRepository struct {
	collection  *mongo.Collection
	logger      *zap.Logger
	transformer domain.MetricsTransformer
}

// NewMetricsRepository creates a new metrics repository
func NewMetricsRepository(collection *mongo.Collection, logger *zap.Logger) domain.MetricsRepository {
	return &metricsRepository{
		collection:  collection,
		logger:      logger,
		transformer: transformers.NewMetricsTransformer(logger),
	}
}

// GetJobMetrics gets the metrics for a job
func (r *metricsRepository) GetJobMetrics(ctx context.Context, jobName string) (database.HostMetrics, error) {
	// Create a filter to find hosts that have service instances with the specified job name
	filter := bson.M{"service_instance_metrics.job_name": jobName}

	// Query the database
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		r.logger.Error("Failed to find metrics from MongoDB", zap.Error(err), zap.String("job_name", jobName))
		return database.HostMetrics{}, err
	}
	defer cursor.Close(ctx)

	// Combine all results into a single DBHostMetrics
	result := database.HostMetrics{
		Host:                   "",
		SystemMetrics:          []database.MetricDatapoints{},
		ServiceInstanceMetrics: []database.ServiceInstanceMetrics{},
	}

	// Process all hosts that match the filter
	for cursor.Next(ctx) {
		var hostMetrics database.HostMetrics
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
		return database.HostMetrics{}, err
	}

	// If no data was found
	if result.Host == "" {
		r.logger.Warn("No metrics found for job", zap.String("job_name", jobName))
		return database.HostMetrics{}, mongo.ErrNoDocuments
	}

	return result, nil
}

func (r *metricsRepository) GetJobMetricsAsMap(ctx context.Context, jobName string) (database.MapHostMetrics, error) {
	metrics, err := r.GetJobMetrics(ctx, jobName)
	if err != nil {
		return database.MapHostMetrics{}, err
	}
	return r.transformer.TransformDBHostMetricsToMap(metrics)
}
