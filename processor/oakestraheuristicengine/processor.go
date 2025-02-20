package oakestraheuristicengine

import (
	"context"
	"fmt"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/constants"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/heuristicentity"
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
	metricStore metricstore.MetricStore

	// policies
	policies              map[string]interfaces.Policy
	policyToEngineMapping map[string]interfaces.HeuristicEntity

	// registry of notification interfaces
	notificationInterfaceRegistry interfaces.NotificationInterfaceRegistry

	// collection of active entities
	activeEntities map[types.HeuristicType]interfaces.HeuristicEntity
	// reponsible to store available entities, needed for initialization
	availableEntities []types.HeuristicType
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
	}, nil
}

// ConsumeMetrics is called when the processor receives metrics
// it saves the metrics to history for later use
func (p *heuristicEngineProcessor) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
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
	notificationInterfaceFactory := policyBuilder.NotificationInterfaceFactory()
	buildInterface := notificationInterfaceFactory.CreateNotificationInterfaceBuilder

	// define notifiers
	routeNotifier := buildInterface(constants.NotificationInterfaceCapability_Route).
		WithHost("localhost").
		WithPort(8080).
		WithEndpoint("/route").
		Build()

	alertNotifier := buildInterface(constants.NotificationInterfaceCapability_Alert).
		WithHost("localhost").
		WithPort(8080).
		WithEndpoint("/alert").
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
			name:             "routing-policy",
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
			WithRouteCondition("true").
			Build()

		// Register policy and its engine mapping
		p.policies[policy.Name()] = policy
		p.policyToEngineMapping[policy.Name()] = engine
	}

	return nil
}

func (p *heuristicEngineProcessor) Shutdown(_ context.Context) error {
	for _, entity := range p.activeEntities {
		if err := entity.Shutdown(); err != nil {
			return err
		}
	}
	return nil
}
