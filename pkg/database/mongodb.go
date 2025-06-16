package database

import (
	"context"
	"fmt"
	"time"

	"github.com/smnzlnsk/opentelemetry-components/pkg/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.uber.org/zap"
)

// Client represents a MongoDB client with connection to a specific database
type MongoDBClient struct {
	client   *mongo.Client
	database *mongo.Database
	logger   *zap.Logger
}

// NewClient creates a new MongoDB client
func NewMongoDBClient(cfg *config.MongoDBConfig, logger *zap.Logger) (*MongoDBClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Set up MongoDB client options with connection pooling optimizations
	clientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%d", cfg.Host, cfg.Port))

	// Set authentication credentials if provided
	if cfg.User != "" && cfg.Password != "" {
		clientOptions.SetAuth(options.Credential{
			Username: cfg.User,
			Password: cfg.Password,
		})
	}

	// Configure connection pooling for better performance
	clientOptions.SetMaxPoolSize(100)                        // Maximum number of connections in the pool
	clientOptions.SetMinPoolSize(10)                         // Minimum number of connections to maintain
	clientOptions.SetMaxConnIdleTime(30 * time.Second)       // Close connections idle for 30 seconds
	clientOptions.SetConnectTimeout(10 * time.Second)        // Connection timeout
	clientOptions.SetSocketTimeout(30 * time.Second)         // Socket timeout for operations
	clientOptions.SetServerSelectionTimeout(5 * time.Second) // Server selection timeout

	// Enable write concern for better performance
	clientOptions.SetWriteConcern(writeconcern.Majority()) // Use majority write concern for consistency

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	// Connect to the database
	database := client.Database("monitoring")

	logger.Info("Connected to MongoDB",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", database.Name()))

	return &MongoDBClient{
		client:   client,
		database: database,
		logger:   logger,
	}, nil
}

// GetDatabase returns the MongoDB database
func (c *MongoDBClient) GetDatabase() *mongo.Database {
	return c.database
}

// Close closes the MongoDB connection
func (c *MongoDBClient) Close(ctx context.Context) error {
	c.logger.Info("Closing MongoDB connection")
	return c.client.Disconnect(ctx)
}
