package oakestraheuristicengine

import (
	"context"
	"fmt"
	"time"

	"github.com/smnzlnsk/opentelemetry-components/pkg/database"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/heuristicentity"
	internalhttp "github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/http"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/logger"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/notification_interface"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/policy"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/repository"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/service"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
)

type heuristicEngineProcessor struct {
	config       *Config
	nextConsumer consumer.Metrics
	logger       *zap.Logger

	// database client
	mongodbClient *database.MongoDBClient
	repositories  *repository.Repositories

	// services
	services domain.Services

	// policies
	policies map[string]domain.Policy

	// collection of active entities
	activeEntities map[domain.HeuristicType]domain.HeuristicEntity
	// reponsible to store available entities, needed for initialization
	availableEntities []domain.HeuristicType

	// http server
	httpServer *internalhttp.Server
}

func newProcessor(config *Config, set processor.Settings, next consumer.Metrics) (*heuristicEngineProcessor, error) {
	return &heuristicEngineProcessor{
		config:            config,
		nextConsumer:      next,
		logger:            set.Logger,
		policies:          make(map[string]domain.Policy),
		activeEntities:    make(map[domain.HeuristicType]domain.HeuristicEntity),
		availableEntities: []domain.HeuristicType{},

		// created on Start
		httpServer:    nil,
		mongodbClient: nil,
		repositories:  nil,
	}, nil
}

// ConsumeMetrics is called when the processor receives metrics
func (p *heuristicEngineProcessor) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	return p.nextConsumer.ConsumeMetrics(ctx, md)
}

func (p *heuristicEngineProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (p *heuristicEngineProcessor) Start(_ context.Context, _ component.Host) error {
	// initialize csv logger
	if err := logger.InitCSVLogger("heuristic_notifications.csv"); err != nil {
		return err
	}

	// initialize mongodb client
	dbClient, err := database.NewMongoDBClient(&p.config.MongoDB, p.logger)
	if err != nil {
		return err
	}

	p.mongodbClient = dbClient
	p.repositories = repository.NewRepositories(dbClient, p.logger)

	// initialize services
	p.services = service.NewServices(p.repositories, p.logger)

	// initialize indexes
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := p.services.GetMetricsService().EnsureIndexes(ctx); err != nil {
			p.logger.Error("Failed to create database indexes", zap.Error(err))
		}
	}()

	for _, entity := range p.activeEntities {
		if err := entity.Start(); err != nil {
			return err
		}
	}

	// TODO: add more entities here
	availableEntities := []domain.HeuristicType{
		domain.RoutingEntity,
	}
	// initialize entity factory
	entityFactory := heuristicentity.NewHeuristicEntityFactory(p.logger)

	// initialize active entities
	for _, entityType := range availableEntities {
		entity, err := entityFactory.CreateHeuristicEntity(entityType, p.services)
		if err != nil {
			return err
		}
		p.activeEntities[entityType] = entity
	}

	// initialize policies
	policyBuilder := policy.NewPolicyBuilder()

	routingNotificationInterfaceBuilder := notification_interface.NewNotificationInterfaceBuilder[any]()

	// define notifiers
	alertNotifier := routingNotificationInterfaceBuilder.
		WithHost(p.config.ServiceManager.Host).
		WithPort(p.config.ServiceManager.Port).
		WithEndpoint("/api/net/routing/alert").
		WithCapability(domain.NotificationInterfaceCapability_Alert).
		Build()

	routingNotifier := routingNotificationInterfaceBuilder.
		WithHost(p.config.ServiceManager.Host).
		WithPort(p.config.ServiceManager.Port).
		WithEndpoint("/api/net/routing/update").
		WithCapability(domain.NotificationInterfaceCapability_Route).
		Build()

	// Verify that the routing heuristic engine exists
	entity, exists := p.activeEntities[domain.RoutingEntity]
	if !exists {
		return fmt.Errorf("heuristic entity %v not found for policy %s", domain.RoutingEntity, "routing")
	}

	p.policies["routing"] = policyBuilder.
		WithName("routing").
		WithHeuristicEntity(entity).
		WithAlertInterface(alertNotifier).
		WithRouteInterface(routingNotifier).
		Build()

	// setup http server if enabled
	if p.config.HTTPServer.Enabled {
		serverConfig := internalhttp.ServerConfig{
			Host: p.config.HTTPServer.Host,
			Port: p.config.HTTPServer.Port,
		}
		p.httpServer = internalhttp.NewServer(serverConfig, p.logger, p.policies, p.services.GetMetricsService())
		if err := p.httpServer.Start(); err != nil {
			p.logger.Error("Failed to start HTTP server", zap.Error(err))
			return err
		}
	}

	p.logger.Info("Processor started")
	return nil
}

func (p *heuristicEngineProcessor) Shutdown(ctx context.Context) error {
	var shutdownErrs []error

	// First shut down the HTTP server if it exists
	if p.httpServer != nil {
		// Create a timeout context specifically for HTTP server shutdown
		httpCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		if err := p.httpServer.Shutdown(httpCtx); err != nil {
			p.logger.Error("Failed to gracefully shut down HTTP server", zap.Error(err))
			shutdownErrs = append(shutdownErrs, fmt.Errorf("HTTP server shutdown: %w", err))
			// Continue with shutdown even if HTTP server shutdown fails
		} else {
			p.logger.Info("HTTP server shutdown successful")
		}
	}

	// Then shut down all entities
	for entityType, entity := range p.activeEntities {
		if err := entity.Shutdown(); err != nil {
			p.logger.Error("Failed to shut down entity", zap.String("entityType", string(entityType)), zap.Error(err))
			shutdownErrs = append(shutdownErrs, fmt.Errorf("entity %s shutdown: %w", entityType, err))
			// Continue with other shutdowns
		}
	}

	// Close MongoDB client
	if p.mongodbClient != nil {
		if err := p.mongodbClient.Close(ctx); err != nil {
			p.logger.Error("Failed to close MongoDB client", zap.Error(err))
			shutdownErrs = append(shutdownErrs, fmt.Errorf("MongoDB client close: %w", err))
		}
	}

	// If we had any errors during shutdown, return a combined error
	if len(shutdownErrs) > 0 {
		return fmt.Errorf("shutdown encountered %d errors: %v", len(shutdownErrs), shutdownErrs)
	}

	p.logger.Info("Processor shutdown complete")
	return nil
}
