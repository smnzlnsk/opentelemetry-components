package oakestraheuristicengine

import (
	"context"
	"fmt"
	"time"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/heuristicentity"
	internalhttp "github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/http"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/notification_interface"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/persistence/mongodb"
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
	mongodbClient *mongodb.Client
	repositories  *repository.Repositories

	// services
	services domain.Services

	// policies
	policies              map[string]domain.Policy
	policyToEngineMapping map[string]domain.HeuristicEntity

	// registry of notification interfaces
	notificationInterfaceRegistry domain.NotificationInterfaceRegistry

	// collection of active entities
	activeEntities map[domain.HeuristicType]domain.HeuristicEntity
	// reponsible to store available entities, needed for initialization
	availableEntities []domain.HeuristicType

	// http server
	httpServer *internalhttp.Server
}

func newProcessor(config *Config, set processor.Settings, next consumer.Metrics) (*heuristicEngineProcessor, error) {
	return &heuristicEngineProcessor{
		config:                        config,
		nextConsumer:                  next,
		logger:                        set.Logger,
		policies:                      make(map[string]domain.Policy),
		policyToEngineMapping:         make(map[string]domain.HeuristicEntity),
		notificationInterfaceRegistry: notification_interface.NewNotificationInterfaceRegistry(set.Logger),
		activeEntities:                make(map[domain.HeuristicType]domain.HeuristicEntity),
		availableEntities:             []domain.HeuristicType{},

		// created on Start
		httpServer:    nil,
		mongodbClient: nil,
		repositories:  nil,
	}, nil
}

// ConsumeMetrics is called when the processor receives metrics
// it saves the metrics to history for later use
func (p *heuristicEngineProcessor) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {

	// save metrics to database
	err := p.services.GetMetricsService().SaveMetrics(ctx, md)
	if err != nil {
		p.logger.Error("failed to save metrics to database", zap.Error(err))
	}

	return p.nextConsumer.ConsumeMetrics(ctx, md)
}

func (p *heuristicEngineProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (p *heuristicEngineProcessor) Start(_ context.Context, _ component.Host) error {
	// initialize mongodb client
	dbClient, err := mongodb.NewClient(&p.config.MongoDB, p.logger)
	if err != nil {
		return err
	}

	p.mongodbClient = dbClient
	p.repositories = repository.NewRepositories(dbClient, p.logger)

	// initialize services
	p.services = service.NewServices(p.repositories, p.logger)

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
	notificationInterfaceBuilder := notification_interface.NewNotificationInterfaceBuilder()

	// define notifiers
	/*routeNotifier := notificationInterfaceBuilder.
	WithHost("localhost").
	WithPort(8080).
	WithEndpoint("/route").
	WithCapability(domain.NotificationInterfaceCapability_Route).
	Build()*/

	alertNotifier := notificationInterfaceBuilder.
		WithHost("localhost").
		WithPort(p.config.ServiceManager.Port).
		WithEndpoint("/api/net/routing/alert").
		WithCapability(domain.NotificationInterfaceCapability_Alert).
		Build()

	routingNotifier := notificationInterfaceBuilder.
		WithHost("localhost").
		WithPort(p.config.ServiceManager.Port).
		WithEndpoint("/api/net/routing/update").
		WithCapability(domain.NotificationInterfaceCapability_Route).
		Build()

	// Define policy configurations with their associated heuristic engine types
	policyConfigs := []struct {
		name             string
		engineType       domain.HeuristicType
		alertNotifier    domain.NotificationInterface
		routeNotifier    domain.NotificationInterface
		scheduleNotifier domain.NotificationInterface
	}{
		{
			name:             "routing",
			engineType:       domain.RoutingEntity,
			alertNotifier:    alertNotifier,
			routeNotifier:    routingNotifier,
			scheduleNotifier: nil,
		},
		// Add more policy configurations here
	}

	// Build and register policies with their associated engines
	for _, cfg := range policyConfigs {
		// Verify that the heuristic engine exists
		engine, exists := p.activeEntities[cfg.engineType]
		if !exists {
			return fmt.Errorf("heuristic engine %v not found for policy %s", cfg.engineType, cfg.name)
		}

		// Build policy with its notifiers and associated engine
		policy := policyBuilder.
			WithName(cfg.name).
			WithPreEvaluationCondition("true").
			WithEvaluationCondition("true").
			WithHeuristicEngine(engine).
			WithAlert(cfg.alertNotifier).
			WithAlertCondition("true").
			WithRoute(cfg.routeNotifier).
			WithRouteCondition("true").
			Build()

		// Register policy and its engine mapping
		p.policies[policy.Name()] = policy
		p.policyToEngineMapping[policy.Name()] = engine
	}

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
