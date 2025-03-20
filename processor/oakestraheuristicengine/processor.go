package oakestraheuristicengine

import (
	"context"
	"fmt"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/constants"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/heuristicentity"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/http"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/metricstore"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/notification_interface"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/policy"
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

	// reponsible to store metrics
	metricStore interfaces.MetricStore

	// policies
	policies              map[string]interfaces.Policy
	policyToEngineMapping map[string]interfaces.HeuristicEntity

	// registry of notification interfaces
	notificationInterfaceRegistry interfaces.NotificationInterfaceRegistry

	// collection of active entities
	activeEntities map[types.HeuristicType]interfaces.HeuristicEntity
	// reponsible to store available entities, needed for initialization
	availableEntities []types.HeuristicType

	// http server
	httpServer *http.Server
}

func newProcessor(config *Config, set processor.Settings, next consumer.Metrics) (*heuristicEngineProcessor, error) {
	// TODO: add more entities here
	availableEntities := []types.HeuristicType{
		constants.RoutingEntity,
	}
	// initialize entity factory
	entityFactory := heuristicentity.NewHeuristicEntityFactory(set.Logger)

	// initialize active entities
	activeEntities := make(map[types.HeuristicType]interfaces.HeuristicEntity)
	for _, entityType := range availableEntities {
		entity, err := entityFactory.CreateHeuristicEntity(entityType)
		if err != nil {
			return nil, err
		}
		activeEntities[entityType] = entity
	}

	return &heuristicEngineProcessor{
		config:                        config,
		nextConsumer:                  next,
		logger:                        set.Logger,
		metricStore:                   metricstore.NewMetricStore(set.Logger),
		policies:                      make(map[string]interfaces.Policy),
		policyToEngineMapping:         make(map[string]interfaces.HeuristicEntity),
		notificationInterfaceRegistry: notification_interface.NewNotificationInterfaceRegistry(set.Logger),
		activeEntities:                activeEntities,
		availableEntities:             availableEntities,
		httpServer:                    nil,
	}, nil
}

// ConsumeMetrics is called when the processor receives metrics
// it saves the metrics to history for later use
func (p *heuristicEngineProcessor) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	err := p.metricStore.Save(md)
	if err != nil {
		p.logger.Error("failed to save metrics to metric store", zap.Error(err))
	}

	// values := p.metricStore.GetValueMapByString()

	/*for _, policy := range p.policies {
		err := policy.Check(values)
		if err == nil {
			// If check passes, enforce the policy which will trigger notifications
			processors := policy.HeuristicEngine().Processors()
			for processorIdentifier := range processors {
				err = policy.Enforce(processorIdentifier, values)
				if err != nil {
					p.logger.Error("failed to enforce policy", zap.String("policy", policy.Name()), zap.Error(err))
				}
			}
		}
	}*/
	return p.nextConsumer.ConsumeMetrics(ctx, md)
}

func (p *heuristicEngineProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (p *heuristicEngineProcessor) Start(_ context.Context, _ component.Host) error {
	for _, entity := range p.activeEntities {
		if err := entity.Start(); err != nil {
			return err
		}
	}

	// initialize policies
	policyBuilder := policy.NewPolicyBuilder()
	notificationInterfaceBuilder := notification_interface.NewNotificationInterfaceBuilder()

	// define notifiers
	routeNotifier := notificationInterfaceBuilder.
		WithHost("localhost").
		WithPort(8080).
		WithEndpoint("/route").
		WithCapability(constants.NotificationInterfaceCapability_Route).
		Build()

	alertNotifier := notificationInterfaceBuilder.
		WithHost("localhost").
		WithPort(8080).
		WithEndpoint("/alert").
		WithCapability(constants.NotificationInterfaceCapability_Alert).
		Build()

	// Define policy configurations with their associated heuristic engine types
	policyConfigs := []struct {
		name             string
		engineType       types.HeuristicType
		alertNotifier    interfaces.NotificationInterface
		routeNotifier    interfaces.NotificationInterface
		scheduleNotifier interfaces.NotificationInterface
	}{
		{
			name:             "routing",
			engineType:       constants.RoutingEntity,
			alertNotifier:    alertNotifier,
			routeNotifier:    routeNotifier,
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
			WithRouteCondition("false").
			Build()

		// Register policy and its engine mapping
		p.policies[policy.Name()] = policy
		p.policyToEngineMapping[policy.Name()] = engine
	}

	// setup http server if enabled
	if p.config.HTTPServer.Enabled {
		serverConfig := http.ServerConfig{
			Host: p.config.HTTPServer.Host,
			Port: p.config.HTTPServer.Port,
		}
		p.httpServer = http.NewServer(serverConfig, p.logger, p.policies, p.metricStore)
		if err := p.httpServer.Start(); err != nil {
			p.logger.Error("Failed to start HTTP server", zap.Error(err))
			return err
		}
		p.logger.Info("Started HTTP server",
			zap.String("host", p.config.HTTPServer.Host),
			zap.Int("port", p.config.HTTPServer.Port))
	}

	return nil
}

func (p *heuristicEngineProcessor) Shutdown(ctx context.Context) error {
	// First shut down the HTTP server if it exists
	if p.httpServer != nil {
		if err := p.httpServer.Shutdown(ctx); err != nil {
			p.logger.Error("Failed to shut down HTTP server", zap.Error(err))
			// Continue with shutdown even if HTTP server shutdown fails
		}
	}

	// Then shut down all entities
	for _, entity := range p.activeEntities {
		if err := entity.Shutdown(); err != nil {
			return err
		}
	}
	return nil
}
