package repository

import (
	"context"
	"fmt"

	"github.com/smnzlnsk/opentelemetry-components/pkg/database"
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

// SaveMetrics saves OpenTelemetry metrics to MongoDB efficiently
// We assume all metrics in the input are from a single host
func (r *metricsRepository) SaveMetrics(ctx context.Context, hostMetrics database.HostMetrics) error {
	// The core issue is that we should not create dynamic field names with dots at all.
	// Instead, use the existing document structure and leverage MongoDB's array update capabilities.

	// For performance, use a more sophisticated approach:
	// 1. Keep the existing document structure (HostMetrics)
	// 2. Use atomic operations to update arrays efficiently
	// 3. Avoid deep dot notation in field names entirely

	filter := bson.M{"host": hostMetrics.Host}

	// Check if document exists first to determine strategy
	var existingDoc database.HostMetrics
	err := r.collection.FindOne(ctx, filter).Decode(&existingDoc)

	if err == mongo.ErrNoDocuments {
		// Insert new document directly - most efficient for new hosts
		_, err = r.collection.InsertOne(ctx, hostMetrics)
		if err != nil {
			r.logger.Error("Failed to insert new host metrics", zap.Error(err), zap.String("host", hostMetrics.Host))
			return err
		}
		return nil
	} else if err != nil {
		r.logger.Error("Failed to query existing host metrics", zap.Error(err), zap.String("host", hostMetrics.Host))
		return err
	}

	// Document exists - merge efficiently using application logic rather than complex MongoDB operations
	// This avoids the dot notation issue entirely and gives us more control
	merged := r.mergeHostMetricsEfficiently(existingDoc, hostMetrics)

	// Replace the entire document (more predictable than complex update operations)
	_, err = r.collection.ReplaceOne(ctx, filter, merged)
	if err != nil {
		r.logger.Error("Failed to replace host metrics", zap.Error(err), zap.String("host", hostMetrics.Host))
		return err
	}

	return nil
}

// mergeHostMetricsEfficiently merges metrics without creating excessive data structures
func (r *metricsRepository) mergeHostMetricsEfficiently(existing, new database.HostMetrics) database.HostMetrics {
	result := existing

	// Merge system metrics efficiently
	result.SystemMetrics = r.mergeMetricDatapoints(result.SystemMetrics, new.SystemMetrics)

	// Merge service instance metrics efficiently
	for _, newService := range new.ServiceInstanceMetrics {
		found := false
		for i, existingService := range result.ServiceInstanceMetrics {
			if existingService.JobName == newService.JobName &&
				existingService.InstanceNumber == newService.InstanceNumber {
				// Merge metrics for existing service instance
				result.ServiceInstanceMetrics[i].Metrics = r.mergeMetricDatapoints(
					result.ServiceInstanceMetrics[i].Metrics,
					newService.Metrics,
				)
				found = true
				break
			}
		}

		if !found {
			// Add new service instance, but limit array size to prevent excessive growth
			result.ServiceInstanceMetrics = append(result.ServiceInstanceMetrics, newService)

			// Keep only recent service instances (prevent excessive growth)
			if len(result.ServiceInstanceMetrics) > 100 {
				result.ServiceInstanceMetrics = result.ServiceInstanceMetrics[len(result.ServiceInstanceMetrics)-100:]
			}
		}
	}

	return result
}

// mergeMetricDatapoints efficiently merges metric datapoints with size limits
func (r *metricsRepository) mergeMetricDatapoints(existing, new []database.MetricDatapoints) []database.MetricDatapoints {
	// Create map for fast lookup
	metricMap := make(map[string]int)

	for i, dp := range existing {
		// Use a simple key that doesn't involve complex dot notation
		key := fmt.Sprintf("%s|%s|%s", dp.Identifier.Name, dp.Identifier.State, dp.Identifier.Type)
		metricMap[key] = i
	}

	// Merge new datapoints
	for _, newDP := range new {
		key := fmt.Sprintf("%s|%s|%s", newDP.Identifier.Name, newDP.Identifier.State, newDP.Identifier.Type)

		if idx, found := metricMap[key]; found {
			// Append new datapoints and limit size
			existing[idx].Datapoints = append(existing[idx].Datapoints, newDP.Datapoints...)

			// Keep only the latest 5 datapoints per metric to prevent excessive growth
			if len(existing[idx].Datapoints) > 5 {
				existing[idx].Datapoints = existing[idx].Datapoints[len(existing[idx].Datapoints)-5:]
			}
		} else {
			// Add new metric, but limit initial datapoints
			if len(newDP.Datapoints) > 5 {
				newDP.Datapoints = newDP.Datapoints[len(newDP.Datapoints)-5:]
			}
			existing = append(existing, newDP)
		}
	}

	// Prevent excessive metric entries per host
	if len(existing) > 50 {
		existing = existing[len(existing)-50:]
	}

	return existing
}

// GetJobMetrics gets the metrics for a job with optimized query using projection
func (r *metricsRepository) GetJobMetrics(ctx context.Context, jobName string) (database.HostMetrics, error) {
	// Create a filter to find hosts that have service instances with the specified job name
	filter := bson.M{"service_instance_metrics.job_name": jobName}

	// Use projection to only fetch relevant fields to reduce network overhead
	projection := bson.M{
		"host":           1,
		"system_metrics": 1,
		"service_instance_metrics": bson.M{
			"$elemMatch": bson.M{
				"job_name": jobName,
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

// SaveMetricsBatch saves multiple host metrics in a single batch operation
func (r *metricsRepository) SaveMetricsBatch(ctx context.Context, hostMetricsList []database.HostMetrics) error {
	if len(hostMetricsList) == 0 {
		return nil
	}

	// Group by host to avoid duplicate operations
	hostMap := make(map[string]database.HostMetrics)
	hostNames := make([]string, 0)

	for _, hostMetrics := range hostMetricsList {
		if existing, exists := hostMap[hostMetrics.Host]; exists {
			// Merge metrics for the same host
			merged := r.mergeHostMetricsEfficiently(existing, hostMetrics)
			hostMap[hostMetrics.Host] = merged
		} else {
			hostMap[hostMetrics.Host] = hostMetrics
			hostNames = append(hostNames, hostMetrics.Host)
		}
	}

	// Fetch all existing documents in a single query for efficiency
	existingDocsMap := make(map[string]database.HostMetrics)
	if len(hostNames) > 0 {
		filter := bson.M{"host": bson.M{"$in": hostNames}}
		cursor, err := r.collection.Find(ctx, filter)
		if err != nil {
			r.logger.Error("Failed to fetch existing documents for batch merge", zap.Error(err))
			return err
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var existingDoc database.HostMetrics
			if err := cursor.Decode(&existingDoc); err != nil {
				r.logger.Error("Failed to decode existing document", zap.Error(err))
				continue
			}
			existingDocsMap[existingDoc.Host] = existingDoc
		}

		if err := cursor.Err(); err != nil {
			r.logger.Error("Cursor error while fetching existing documents", zap.Error(err))
			return err
		}
	}

	// Create bulk write operations that preserve historic data
	var operations []mongo.WriteModel

	for host, newMetrics := range hostMap {
		filter := bson.M{"host": host}

		var finalMetrics database.HostMetrics
		if existingDoc, exists := existingDocsMap[host]; exists {
			// Merge with existing document to preserve historic data
			finalMetrics = r.mergeHostMetricsEfficiently(existingDoc, newMetrics)
		} else {
			// No existing document, use new metrics as-is
			finalMetrics = newMetrics
		}

		// Create replace operation with merged data
		replaceModel := mongo.NewReplaceOneModel()
		replaceModel.SetFilter(filter)
		replaceModel.SetReplacement(finalMetrics)
		replaceModel.SetUpsert(true)

		operations = append(operations, replaceModel)
	}

	if len(operations) == 0 {
		r.logger.Warn("No valid operations to execute in batch")
		return nil
	}

	// Execute bulk write
	opts := options.BulkWrite().SetOrdered(false) // Allow parallel execution
	result, err := r.collection.BulkWrite(ctx, operations, opts)
	if err != nil {
		r.logger.Error("Failed to execute bulk write", zap.Error(err))
		return err
	}

	r.logger.Debug("Bulk write completed with historic data preservation",
		zap.Int("inserted", int(result.InsertedCount)),
		zap.Int("modified", int(result.ModifiedCount)),
		zap.Int("upserted", int(result.UpsertedCount)),
		zap.Int("hosts_processed", len(hostMap)))

	return nil
}
