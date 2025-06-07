package database

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// DatabaseConfig holds configuration for any database type
type DatabaseConfig struct {
	MongoDB *MongoDBConfig `yaml:"mongodb,omitempty"`
	Redis   *RedisConfig   `yaml:"redis,omitempty"`
}

// ClientFactory creates database clients based on configuration
type ClientFactory struct {
	logger *zap.Logger
}

// NewClientFactory creates a new client factory
func NewClientFactory(logger *zap.Logger) *ClientFactory {
	return &ClientFactory{
		logger: logger,
	}
}

// CreateClient creates a database client based on the configuration
func (f *ClientFactory) CreateClient(cfg *DatabaseConfig) (Client, error) {
	// Determine database type from which configuration is present
	if cfg.MongoDB != nil && cfg.Redis != nil {
		return nil, fmt.Errorf("only one database configuration can be specified")
	}

	if cfg.MongoDB != nil {
		return NewMongoDBClient(cfg.MongoDB, f.logger)
	}

	if cfg.Redis != nil {
		return NewRedisClient(cfg.Redis, f.logger)
	}

	return nil, fmt.Errorf("no database configuration specified - either 'mongodb' or 'redis' is required")
}

// CreateAndConnect creates and connects a database client
func (f *ClientFactory) CreateAndConnect(ctx context.Context, cfg *DatabaseConfig) (Client, error) {
	client, err := f.CreateClient(cfg)
	if err != nil {
		return nil, err
	}

	if err := client.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return client, nil
}
