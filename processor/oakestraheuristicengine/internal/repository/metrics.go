package repository

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/middleware"
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
		transformer: middleware.NewMetricsTransformer(logger),
	}
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
		SystemMetrics:          []domain.DBMetricDatapoint{},
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

func (r *metricsRepository) GetJobMetricsAsMap(ctx context.Context, jobName string) (domain.MapHostMetrics, error) {
	metrics, err := r.GetJobMetrics(ctx, jobName)
	if err != nil {
		return domain.MapHostMetrics{}, err
	}
	return r.transformer.TransformDBHostMetricsToMap(metrics)
}
