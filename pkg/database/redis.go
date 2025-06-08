package database

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisClient represents a Redis client implementation
type RedisClient struct {
	client       *redis.Client
	logger       *zap.Logger
	config       *RedisConfig
	metricsStore MetricsStore
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// Ensure RedisClient implements the Client interface
var _ Client = (*RedisClient)(nil)

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg *RedisConfig, logger *zap.Logger) (*RedisClient, error) {
	client := &RedisClient{
		config: cfg,
		logger: logger,
	}

	return client, nil
}

// Connect establishes a connection to Redis
func (c *RedisClient) Connect(ctx context.Context) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.config.Host, c.config.Port),
		Password: c.config.Password,
		DB:       c.config.DB,
	})

	// Test the connection
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	c.client = rdb
	c.metricsStore = NewRedisMetricsStore(rdb, c.logger)

	c.logger.Info("Connected to Redis",
		zap.String("host", c.config.Host),
		zap.Int("port", c.config.Port),
		zap.Int("db", c.config.DB))

	return nil
}

// Health checks if the Redis connection is healthy
func (c *RedisClient) Health(ctx context.Context) error {
	if c.client == nil {
		return fmt.Errorf("redis client not connected")
	}
	_, err := c.client.Ping(ctx).Result()
	return err
}

// GetMetricsStore returns the metrics store
func (c *RedisClient) GetMetricsStore() MetricsStore {
	return c.metricsStore
}

// Close closes the Redis connection
func (c *RedisClient) Close(ctx context.Context) error {
	c.logger.Info("Closing Redis connection")
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// redisMetricsStore implements the MetricsStore interface for Redis
type redisMetricsStore struct {
	client *redis.Client
	logger *zap.Logger
}

// Ensure redisMetricsStore implements the MetricsStore interface
var _ MetricsStore = (*redisMetricsStore)(nil)

// NewRedisMetricsStore creates a new Redis metrics store
func NewRedisMetricsStore(client *redis.Client, logger *zap.Logger) MetricsStore {
	return &redisMetricsStore{
		client: client,
		logger: logger,
	}
}

// SaveMetrics saves host metrics to Redis
func (s *redisMetricsStore) SaveMetrics(ctx context.Context, metrics HostMetrics) error {
	// Use the host as the key with a timestamp
	key := fmt.Sprintf("metrics:%s:%d", metrics.Host, time.Now().Unix())

	// Convert metrics to JSON
	data, err := json.Marshal(metrics)
	if err != nil {
		s.logger.Error("Failed to marshal metrics", zap.Error(err))
		return err
	}

	// Set the data in Redis with a TTL of 24 hours
	err = s.client.Set(ctx, key, data, 24*time.Hour).Err()
	if err != nil {
		s.logger.Error("Failed to save metrics to Redis", zap.Error(err))
		return err
	}

	// Also maintain a set of job names for efficient lookup
	for _, serviceInstance := range metrics.ServiceInstanceMetrics {
		jobKey := fmt.Sprintf("job:%s", serviceInstance.JobName)
		err = s.client.SAdd(ctx, jobKey, key).Err()
		if err != nil {
			s.logger.Error("Failed to add job reference to Redis", zap.Error(err))
		}
		// Set TTL for job key as well
		s.client.Expire(ctx, jobKey, 24*time.Hour)
	}

	return nil
}

// GetJobMetrics retrieves metrics for a specific job from Redis
func (s *redisMetricsStore) GetJobMetrics(ctx context.Context, jobName string) (HostMetrics, error) {
	jobKey := fmt.Sprintf("job:%s", jobName)

	// Get all metric keys for this job
	metricKeys, err := s.client.SMembers(ctx, jobKey).Result()
	if err != nil {
		s.logger.Error("Failed to get job metric keys from Redis", zap.Error(err))
		return HostMetrics{}, err
	}

	if len(metricKeys) == 0 {
		s.logger.Warn("No metrics found for job", zap.String("job_name", jobName))
		return HostMetrics{}, fmt.Errorf("no metrics found for job: %s", jobName)
	}

	// Combine all results into a single HostMetrics
	result := HostMetrics{
		Host:                   "",
		SystemMetrics:          []MetricDatapoints{},
		ServiceInstanceMetrics: []ServiceInstanceMetrics{},
	}

	// Retrieve and combine all metrics
	for _, key := range metricKeys {
		data, err := s.client.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				// Key expired or doesn't exist, clean up the reference
				s.client.SRem(ctx, jobKey, key)
				continue
			}
			s.logger.Error("Failed to get metric data from Redis", zap.Error(err), zap.String("key", key))
			continue
		}

		var hostMetrics HostMetrics
		if err := json.Unmarshal([]byte(data), &hostMetrics); err != nil {
			s.logger.Error("Failed to unmarshal metric data", zap.Error(err), zap.String("key", key))
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

	// If no data was found
	if result.Host == "" {
		s.logger.Warn("No valid metrics found for job", zap.String("job_name", jobName))
		return HostMetrics{}, fmt.Errorf("no valid metrics found for job: %s", jobName)
	}

	return result, nil
}

// GetJobMetricsAsMap retrieves metrics for a specific job as a map from Redis
func (s *redisMetricsStore) GetJobMetricsAsMap(ctx context.Context, jobName string) (HostMetricsMap, error) {
	metrics, err := s.GetJobMetrics(ctx, jobName)
	if err != nil {
		return HostMetricsMap{}, err
	}

	// Transform the metrics to a map format
	return s.transformHostMetricsToMap(metrics)
}

// DeleteJobMetrics deletes metrics for a specific job from Redis
func (s *redisMetricsStore) DeleteJobMetrics(ctx context.Context, jobName string) error {
	jobKey := fmt.Sprintf("job:%s", jobName)

	// Get all metric keys for this job
	metricKeys, err := s.client.SMembers(ctx, jobKey).Result()
	if err != nil {
		s.logger.Error("Failed to get job metric keys for deletion", zap.Error(err))
		return err
	}

	// Delete all metric data
	if len(metricKeys) > 0 {
		err = s.client.Del(ctx, metricKeys...).Err()
		if err != nil {
			s.logger.Error("Failed to delete job metrics from Redis", zap.Error(err))
			return err
		}
	}

	// Delete the job key itself
	err = s.client.Del(ctx, jobKey).Err()
	if err != nil {
		s.logger.Error("Failed to delete job key from Redis", zap.Error(err))
		return err
	}

	return nil
}

// DeleteExpiredMetrics deletes metrics older than maxAge seconds from Redis
func (s *redisMetricsStore) DeleteExpiredMetrics(ctx context.Context, maxAge int64) error {
	// Redis automatically handles expiration via TTL, but we can also manually clean up
	cutoffTime := time.Now().Add(-time.Duration(maxAge) * time.Second).Unix()

	// Scan for metric keys
	iter := s.client.Scan(ctx, 0, "metrics:*", 0).Iterator()
	var expiredKeys []string

	for iter.Next(ctx) {
		key := iter.Val()

		// Extract timestamp from key (format: metrics:host:timestamp)
		parts := splitKey(key)
		if len(parts) >= 3 {
			if timestamp, err := strconv.ParseInt(parts[2], 10, 64); err == nil {
				if timestamp < cutoffTime {
					expiredKeys = append(expiredKeys, key)
				}
			}
		}
	}

	if err := iter.Err(); err != nil {
		s.logger.Error("Failed to scan for expired metrics", zap.Error(err))
		return err
	}

	// Delete expired keys
	if len(expiredKeys) > 0 {
		err := s.client.Del(ctx, expiredKeys...).Err()
		if err != nil {
			s.logger.Error("Failed to delete expired metrics from Redis", zap.Error(err))
			return err
		}
		s.logger.Info("Deleted expired metrics", zap.Int("count", len(expiredKeys)))
	}

	return nil
}

// transformHostMetricsToMap transforms HostMetrics to HostMetricsMap format
func (s *redisMetricsStore) transformHostMetricsToMap(dbHostMetrics HostMetrics) (HostMetricsMap, error) {
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

func (s *redisMetricsStore) buildMetricID(name string, state string, age int) string {
	return fmt.Sprintf("%s(%d){%s}", name, age, state)
}

func (s *redisMetricsStore) calculateAge(length int, index int) int {
	return (length - 1) - index
}

// Helper function to split Redis keys (since redis package might not have SplitKey)
func splitKey(key string) []string {
	result := []string{}
	current := ""

	for _, char := range key {
		if char == ':' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}
