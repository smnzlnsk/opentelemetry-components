package repository

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/pkg/database"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/transformers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

// EnsureIndexes creates necessary indexes for optimal query performance
func (r *metricsRepository) EnsureIndexes(ctx context.Context) error {
	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "host", Value: 1}},
			Options: options.Index().SetName("host_index").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "service_instance_metrics.job_name", Value: 1}},
			Options: options.Index().SetName("job_name_index"),
		},
		{
			Keys: bson.D{
				{Key: "service_instance_metrics.job_name", Value: 1},
				{Key: "service_instance_metrics.instance_number", Value: 1},
			},
			Options: options.Index().SetName("job_instance_index"),
		},
		{
			Keys:    bson.D{{Key: "system_metrics.identifier.name", Value: 1}},
			Options: options.Index().SetName("system_metric_name_index"),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		r.logger.Error("Failed to create database indexes", zap.Error(err))
		return err
	}

	r.logger.Info("Database indexes created successfully")
	return nil
}

// GetJobMetrics gets the metrics for a job with optimized query using projection
func (r *metricsRepository) GetJobMetrics(ctx context.Context, jobName string) (database.HostMetrics, error) {
	// Create a filter to find hosts that have service instances with the specified job name
	filter := bson.M{"service_instance_metrics.job_name": jobName}

	// Use projection to only fetch relevant fields to reduce network overhead
	// Note: Using $filter instead of $elemMatch to get ALL matching service instances
	projection := bson.M{
		"host":           1,
		"system_metrics": 1,
		"service_instance_metrics": bson.M{
			"$filter": bson.M{
				"input": "$service_instance_metrics",
				"cond":  bson.M{"$eq": bson.A{"$$this.job_name", jobName}},
			},
		},
	}

	// Create find options with projection
	findOptions := options.Find().SetProjection(projection)

	// Query the database with projection
	cursor, err := r.collection.Find(ctx, filter, findOptions)
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

		// Add only service instance metrics that match the jobName (projection should have already filtered)
		result.ServiceInstanceMetrics = append(result.ServiceInstanceMetrics, hostMetrics.ServiceInstanceMetrics...)
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

// GetJobMetricsBatch gets metrics for multiple jobs in a single query
func (r *metricsRepository) GetJobMetricsBatch(ctx context.Context, jobNames []string) (map[string]database.HostMetrics, error) {
	if len(jobNames) == 0 {
		return make(map[string]database.HostMetrics), nil
	}

	// Create a filter to find hosts that have service instances with any of the specified job names
	filter := bson.M{"service_instance_metrics.job_name": bson.M{"$in": jobNames}}

	// Use projection to only fetch relevant fields
	projection := bson.M{
		"host":           1,
		"system_metrics": 1,
		"service_instance_metrics": bson.M{
			"$filter": bson.M{
				"input": "$service_instance_metrics",
				"cond":  bson.M{"$in": bson.A{"$$this.job_name", jobNames}},
			},
		},
	}

	// Create find options with projection
	findOptions := options.Find().SetProjection(projection)

	// Query the database
	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		r.logger.Error("Failed to find batch metrics from MongoDB", zap.Error(err), zap.Strings("job_names", jobNames))
		return nil, err
	}
	defer cursor.Close(ctx)

	// Results map: job_name -> HostMetrics
	results := make(map[string]database.HostMetrics)

	// Initialize results for all requested jobs
	for _, jobName := range jobNames {
		results[jobName] = database.HostMetrics{
			Host:                   "",
			SystemMetrics:          []database.MetricDatapoints{},
			ServiceInstanceMetrics: []database.ServiceInstanceMetrics{},
		}
	}

	// Process all hosts that match the filter
	for cursor.Next(ctx) {
		var hostMetrics database.HostMetrics
		if err := cursor.Decode(&hostMetrics); err != nil {
			r.logger.Error("Failed to decode host metrics in batch", zap.Error(err))
			continue
		}

		// Group service instance metrics by job name
		for _, serviceInstance := range hostMetrics.ServiceInstanceMetrics {
			jobName := serviceInstance.JobName
			if result, exists := results[jobName]; exists {
				// Set host name if not set
				if result.Host == "" {
					result.Host = hostMetrics.Host
					result.SystemMetrics = append(result.SystemMetrics, hostMetrics.SystemMetrics...)
				}
				result.ServiceInstanceMetrics = append(result.ServiceInstanceMetrics, serviceInstance)
				results[jobName] = result
			}
		}
	}

	if err := cursor.Err(); err != nil {
		r.logger.Error("Cursor error while getting batch job metrics", zap.Error(err))
		return nil, err
	}

	return results, nil
}

func (r *metricsRepository) GetJobMetricsAsMap(ctx context.Context, jobName string) (database.HostMetricsMap, error) {
	metrics, err := r.GetJobMetrics(ctx, jobName)
	if err != nil {
		return database.HostMetricsMap{}, err
	}
	return r.transformer.TransformHostMetricsToMap(metrics)
}

// GetJobMetricsAsMapBatch gets metrics as map for multiple jobs in a single operation
func (r *metricsRepository) GetJobMetricsAsMapBatch(ctx context.Context, jobNames []string) (map[string]database.HostMetricsMap, error) {
	batchMetrics, err := r.GetJobMetricsBatch(ctx, jobNames)
	if err != nil {
		return nil, err
	}

	results := make(map[string]database.HostMetricsMap)
	for jobName, metrics := range batchMetrics {
		transformed, err := r.transformer.TransformHostMetricsToMap(metrics)
		if err != nil {
			r.logger.Error("Failed to transform metrics to map", zap.Error(err), zap.String("job_name", jobName))
			continue
		}
		results[jobName] = transformed
	}

	return results, nil
}
