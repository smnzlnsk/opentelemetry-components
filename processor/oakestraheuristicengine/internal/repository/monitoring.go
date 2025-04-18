package repository

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

// monitoringRepository implements domain.MonitoringRepository
// It handles storing OpenTelemetry metrics in MongoDB
type monitoringRepository struct {
	collection  *mongo.Collection
	logger      *zap.Logger
	transformer domain.MetricsTransformer
}

// NewMonitoringRepository creates a new monitoring repository
func NewMonitoringRepository(collection *mongo.Collection, logger *zap.Logger) domain.MonitoringRepository {
	return &monitoringRepository{
		collection:  collection,
		logger:      logger,
		transformer: middleware.NewMetricsTransformer(logger),
	}
}

// SaveMetrics saves OpenTelemetry metrics to MongoDB
// We assume all metrics in the input are from a single host
func (r *monitoringRepository) SaveMetrics(ctx context.Context, md pmetric.Metrics) error {
	// Extract the host from metrics using the transformer
	host := r.transformer.ExtractHost(md)

	// Create the host metrics document using the transformer
	hostMetrics, err := r.transformer.TransformToDBHostMetrics(md)
	if err != nil {
		r.logger.Error("Failed to transform metrics", zap.Error(err))
		return err
	}

	// Create filter for upsert
	filter := bson.M{"host": host}

	// Create update options with upsert
	opts := options.Update().SetUpsert(true)

	// First try to find the existing document
	var existingHost domain.DBHostMetrics
	err = r.collection.FindOne(ctx, filter).Decode(&existingHost)

	if err != nil && err != mongo.ErrNoDocuments {
		r.logger.Error("Failed to query existing host metrics", zap.Error(err), zap.String("host", host))
		return err
	}

	if err == mongo.ErrNoDocuments {
		// Insert new document
		_, err = r.collection.InsertOne(ctx, hostMetrics)
		if err != nil {
			r.logger.Error("Failed to insert host metrics", zap.Error(err), zap.String("host", host))
			return err
		}
	} else {
		// Merge metrics with existing
		merged := r.mergeHostMetrics(existingHost, hostMetrics)

		// Update the document
		update := bson.M{"$set": merged}
		_, err = r.collection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			r.logger.Error("Failed to update host metrics", zap.Error(err), zap.String("host", host))
			return err
		}
	}

	return nil
}

// mergeHostMetrics merges new metrics into existing host metrics
func (r *monitoringRepository) mergeHostMetrics(existing, new domain.DBHostMetrics) domain.DBHostMetrics {
	result := existing

	// Helper function to add new datapoints to existing metrics
	mergeMetricDatapoints := func(existing []domain.DBMetricDatapoint, new []domain.DBMetricDatapoint) []domain.DBMetricDatapoint {
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

				// Keep only the latest N datapoints (e.g., 100)
				if len(existing[idx].Datapoints) > 100 {
					existing[idx].Datapoints = existing[idx].Datapoints[len(existing[idx].Datapoints)-100:]
				}
			} else {
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

// GetHostInstanceMetrics gets the metrics for a host and service instance
func (r *monitoringRepository) GetHostInstanceMetrics(ctx context.Context, host string, serviceInstance string) (domain.DBHostMetrics, error) {
	// Create a filter to find the host
	filter := bson.M{"host": host}

	// Query the database
	var hostMetrics domain.DBHostMetrics
	err := r.collection.FindOne(ctx, filter).Decode(&hostMetrics)
	if err != nil {
		r.logger.Error("Failed to get metrics from MongoDB", zap.Error(err), zap.String("host", host))
		return domain.DBHostMetrics{Host: host}, err
	}

	// If service instance is specified, filter to only that service
	if serviceInstance != "" {
		// Filter to only the requested service
		filteredMetrics := domain.DBHostMetrics{
			Host:                   hostMetrics.Host,
			SystemMetrics:          []domain.DBMetricDatapoint{},
			ServiceInstanceMetrics: []domain.DBServiceInstanceMetrics{},
		}

		// Find and include only the requested service instance
		for _, service := range hostMetrics.ServiceInstanceMetrics {
			if service.JobName == serviceInstance {
				filteredMetrics.ServiceInstanceMetrics = append(filteredMetrics.ServiceInstanceMetrics, service)
				break
			}
		}

		return filteredMetrics, nil
	}

	return hostMetrics, nil
}
