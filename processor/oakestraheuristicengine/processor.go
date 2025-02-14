package oakestraheuristicengine

import (
	"context"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/constants"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/heuristicentity"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/measure"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/metricstore"
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

	measureRegistry interfaces.MeasureRegistry

	// collection of active entities
	entityFactory  interfaces.HeuristicEntityFactory
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
		config:            config,
		nextConsumer:      next,
		logger:            set.Logger,
		metricStore:       metricstore.NewMetricStore(set.Logger),
		measureRegistry:   measure.NewMeasureRegistry(set.Logger),
		entityFactory:     entityFactory,
		activeEntities:    activeEntities,
		availableEntities: availableEntities,
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
	measureFactory := policyBuilder.MeasureFactory()

	// add policies
	policyBuilder.WithName("routing-policy").
		WithRoute(measureFactory.CreateMeasure(constants.MeasureTypeRoute)).
		Build()

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
