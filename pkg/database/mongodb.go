package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// MongoDBClient represents a MongoDB client with connection to a specific database
type MongoDBClient struct {
	client       *mongo.Client
	database     *mongo.Database
	logger       *zap.Logger
	config       *MongoDBConfig
	metricsStore MetricsStore
}

// MongoDBConfig represents MongoDB configuration
type MongoDBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

// Ensure MongoDBClient implements the Client interface
var _ Client = (*MongoDBClient)(nil)

// NewMongoDBClient creates a new MongoDB client
func NewMongoDBClient(cfg *MongoDBConfig, logger *zap.Logger) (*MongoDBClient, error) {
	client := &MongoDBClient{
		config: cfg,
		logger: logger,
	}

	return client, nil
}

// Connect establishes a connection to MongoDB
func (c *MongoDBClient) Connect(ctx context.Context) error {
	// Set up MongoDB client options
	clientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%d", c.config.Host, c.config.Port))

	// Set authentication credentials if provided
	if c.config.User != "" && c.config.Password != "" {
		clientOptions.SetAuth(options.Credential{
			Username: c.config.User,
			Password: c.config.Password,
		})
	}

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return err
	}

	// Connect to the database
	database := client.Database("monitoring")

	c.client = client
	c.database = database
	c.metricsStore = NewMongoMetricsStore(database.Collection("metrics"), c.logger)

	c.logger.Info("Connected to MongoDB",
		zap.String("host", c.config.Host),
		zap.Int("port", c.config.Port),
		zap.String("database", database.Name()))

	return nil
}

// Health checks if the MongoDB connection is healthy
func (c *MongoDBClient) Health(ctx context.Context) error {
	if c.client == nil {
		return fmt.Errorf("mongodb client not connected")
	}
	return c.client.Ping(ctx, nil)
}

// GetMetricsStore returns the metrics store
func (c *MongoDBClient) GetMetricsStore() MetricsStore {
	return c.metricsStore
}

// GetDatabase returns the MongoDB database
func (c *MongoDBClient) GetDatabase() *mongo.Database {
	return c.database
}

// Close closes the MongoDB connection
func (c *MongoDBClient) Close(ctx context.Context) error {
	c.logger.Info("Closing MongoDB connection")
	if c.client != nil {
		return c.client.Disconnect(ctx)
	}
	return nil
}

// mongoMetricsStore implements the MetricsStore interface for MongoDB
type mongoMetricsStore struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

// Ensure mongoMetricsStore implements the MetricsStore interface
var _ MetricsStore = (*mongoMetricsStore)(nil)

// NewMongoMetricsStore creates a new MongoDB metrics store
func NewMongoMetricsStore(collection *mongo.Collection, logger *zap.Logger) MetricsStore {
	return &mongoMetricsStore{
		collection: collection,
		logger:     logger,
	}
}

// SaveMetrics saves host metrics to MongoDB
func (s *mongoMetricsStore) SaveMetrics(ctx context.Context, metrics HostMetrics) error {
	_, err := s.collection.InsertOne(ctx, metrics)
	if err != nil {
		s.logger.Error("Failed to save metrics to MongoDB", zap.Error(err))
		return err
	}
	return nil
}

// GetJobMetrics retrieves metrics for a specific job from MongoDB
func (s *mongoMetricsStore) GetJobMetrics(ctx context.Context, jobName string) (HostMetrics, error) {
	// Create a filter to find hosts that have service instances with the specified job name
	filter := bson.M{"service_instance_metrics.job_name": jobName}

	// Query the database
	cursor, err := s.collection.Find(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to find metrics from MongoDB", zap.Error(err), zap.String("job_name", jobName))
		return HostMetrics{}, err
	}
	defer cursor.Close(ctx)

	// Combine all results into a single HostMetrics
	result := HostMetrics{
		Host:                   "",
		SystemMetrics:          []MetricDatapoints{},
		ServiceInstanceMetrics: []ServiceInstanceMetrics{},
	}

	// Process all hosts that match the filter
	for cursor.Next(ctx) {
		var hostMetrics HostMetrics
		if err := cursor.Decode(&hostMetrics); err != nil {
			s.logger.Error("Failed to decode host metrics", zap.Error(err), zap.String("job_name", jobName))
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
		s.logger.Error("Cursor error while getting job metrics", zap.Error(err), zap.String("job_name", jobName))
		return HostMetrics{}, err
	}

	// If no data was found
	if result.Host == "" {
		s.logger.Warn("No metrics found for job", zap.String("job_name", jobName))
		return HostMetrics{}, mongo.ErrNoDocuments
	}

	return result, nil
}

// GetJobMetricsAsMap retrieves metrics for a specific job as a map from MongoDB
func (s *mongoMetricsStore) GetJobMetricsAsMap(ctx context.Context, jobName string) (HostMetricsMap, error) {
	metrics, err := s.GetJobMetrics(ctx, jobName)
	if err != nil {
		return HostMetricsMap{}, err
	}

	// Transform the metrics to a map format
	return s.transformHostMetricsToMap(metrics)
}

// DeleteJobMetrics deletes metrics for a specific job from MongoDB
func (s *mongoMetricsStore) DeleteJobMetrics(ctx context.Context, jobName string) error {
	filter := bson.M{"service_instance_metrics.job_name": jobName}
	_, err := s.collection.DeleteMany(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to delete job metrics from MongoDB", zap.Error(err), zap.String("job_name", jobName))
		return err
	}
	return nil
}

// DeleteExpiredMetrics deletes metrics older than maxAge seconds from MongoDB
func (s *mongoMetricsStore) DeleteExpiredMetrics(ctx context.Context, maxAge int64) error {
	cutoffTime := time.Now().Add(-time.Duration(maxAge) * time.Second)
	filter := bson.M{
		"$or": []bson.M{
			{"system_metrics.datapoints.timestamp": bson.M{"$lt": cutoffTime}},
			{"service_instance_metrics.metrics.datapoints.timestamp": bson.M{"$lt": cutoffTime}},
		},
	}
	_, err := s.collection.DeleteMany(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to delete expired metrics from MongoDB", zap.Error(err))
		return err
	}
	return nil
}

// transformHostMetricsToMap transforms HostMetrics to HostMetricsMap format
func (s *mongoMetricsStore) transformHostMetricsToMap(dbHostMetrics HostMetrics) (HostMetricsMap, error) {
	hostMetrics := make(HostMetricsMap)

	// Initialize the host entry with empty maps
	hostMetrics[dbHostMetrics.Host] = MetricsMap{
		HostMetrics:            make(map[string]float64),
		ServiceInstanceMetrics: make(map[string]map[string]float64),
	}

	// Process system metrics
	for _, systemMetric := range dbHostMetrics.SystemMetrics {
		for i, datapoint := range systemMetric.Datapoints {
			// Create a unique identifier for the metric
			age := s.calculateAge(len(systemMetric.Datapoints), i)
			metricID := s.buildMetricID(systemMetric.Identifier.Name, systemMetric.Identifier.State, age)
			hostMetrics[dbHostMetrics.Host].HostMetrics[metricID] = datapoint.Value
		}
	}

	// Process service instance metrics
	for _, serviceInstance := range dbHostMetrics.ServiceInstanceMetrics {
		// Create the service identifier
		serviceID := fmt.Sprintf("%s.instance.%d", serviceInstance.JobName, serviceInstance.InstanceNumber)

		// Initialize the service metrics map if it doesn't exist
		if _, exists := hostMetrics[dbHostMetrics.Host].ServiceInstanceMetrics[serviceID]; !exists {
			hostMetrics[dbHostMetrics.Host].ServiceInstanceMetrics[serviceID] = make(map[string]float64)
		}

		// Add each metric datapoint
		for _, metric := range serviceInstance.Metrics {
			for i, datapoint := range metric.Datapoints {
				// Create a unique identifier for the metric
				age := s.calculateAge(len(metric.Datapoints), i)
				metricID := s.buildMetricID(metric.Identifier.Name, metric.Identifier.State, age)
				hostMetrics[dbHostMetrics.Host].ServiceInstanceMetrics[serviceID][metricID] = datapoint.Value
			}
		}
	}

	return hostMetrics, nil
}

func (s *mongoMetricsStore) buildMetricID(name string, state string, age int) string {
	return fmt.Sprintf("%s(%d){%s}", name, age, state)
}

func (s *mongoMetricsStore) calculateAge(length int, index int) int {
	return (length - 1) - index
}
