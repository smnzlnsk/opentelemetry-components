package memoryprocessor

import (
	"context"
	"fmt"
	"time"

	pb "github.com/smnzlnsk/monitoring-proto-lib/gen/go/monitoring_proto_lib/monitoring/v1"
	"github.com/smnzlnsk/opentelemetry-components/internal/shared/calculation"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/processor/memoryprocessor/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

type MemoryMetricProcessor struct {
	contracts *domain.ContractState // create a per-service map of calculation contracts
	config    *Config
	logger    *zap.Logger
	cancel    context.CancelFunc
	settings  processor.Settings
	mb        *metadata.MetricsBuilder
	services  domain.Services
}

var _ internal.MetricProcessor = (*MemoryMetricProcessor)(nil)

func (c *MemoryMetricProcessor) ProcessMetrics(metrics pmetric.Metrics) error {

	m, err := c.processMetrics(metrics)
	if err != nil {
		return err
	}

	m.ResourceMetrics().MoveAndAppendTo(metrics.ResourceMetrics())
	return nil
}

func (c *MemoryMetricProcessor) processMetrics(_ pmetric.Metrics) (pmetric.Metrics, error) {
	// setup new calculation mechanism

	results := c.contracts.Evaluate()

	for key, value := range results {
		rb := c.mb.NewResourceBuilder()
		rb.SetServiceName(key.Service)
		rb.SetContainerID(key.Service)
		c.mb.RecordServiceMemoryUtilisationDataPoint(
			pcommon.NewTimestampFromTime(time.Now()),
			value,
			metadata.MapAttributeState[key.State],
		)

		// set resources
		c.mb.EmitForResource(metadata.WithResource(rb.Emit()))
	}
	return c.mb.Emit(), nil
}

func (c *MemoryMetricProcessor) Shutdown(_ context.Context) error {
	if c.cancel != nil {
		c.cancel()
	}
	c.logger.Info("Stopped Memory Processor")
	return nil
}

func (c *MemoryMetricProcessor) Start(ctx context.Context, _ component.Host) error {
	_, c.cancel = context.WithCancel(ctx)

	// sync contracts
	c.contracts.Sync()

	defaultContracts := []struct {
		Formula string
		States  []string
	}{
		{
			Formula: "([container.memory.usage] / [system.memory.usage]) * 1000000",
			States:  []string{"slab_reclaimable", "slab_unreclaimable", "used"},
		},
	}

	// initialize default contracts
	for _, contract := range defaultContracts {
		if err := c.contracts.GenerateDefaultContract(contract.Formula, contract.States); err != nil {
			return err
		}
	}

	// initialize metric builder
	c.mb = metadata.NewMetricsBuilder(c.config.MetricsBuilderConfig, receiver.Settings{TelemetrySettings: c.settings.TelemetrySettings})
	c.logger.Info("Started Memory Processor")
	return nil
}

func newMemoryMetricProcessor(
	_ context.Context,
	set processor.Settings,
	cfg internal.Config,
	services domain.Services,
	dm domain.DatapointManager,
) (internal.MetricProcessor, error) {
	contracts, err := domain.NewContractState(TypeStr, set.Logger, services, dm)
	if err != nil {
		return nil, err
	}
	return &MemoryMetricProcessor{
		contracts: contracts,
		config:    cfg.(*Config),
		settings:  set,
		logger:    set.Logger,
		services:  services,
	}, nil
}

func (c *MemoryMetricProcessor) RegisterService(serviceName string, instanceNumber int32, resource *pb.ResourceInfo, _ []*pb.CalculationRequest) error {
	// register default services in internal contract state
	err := c.contracts.RegisterService(fmt.Sprintf("%s.instance.%d", serviceName, instanceNumber), []calculation.Contract{}, resource.Memory)
	if err != nil {
		return err
	}

	defContracts := c.contracts.GetDefaultContracts()
	contractsArray := make([]calculation.Contract, 0, len(defContracts))
	// Change service name from default to serviceName
	for _, contract := range defContracts {
		contract.Service = fmt.Sprintf("%s.instance.%d", serviceName, instanceNumber)
		contractsArray = append(contractsArray, contract)
	}

	// notify contract service to create contracts
	ctx := context.Background()
	err = c.services.GetContractService().CreateMany(ctx, contractsArray)
	if err != nil {
		return err
	}

	return nil
}

func (c *MemoryMetricProcessor) DeleteService(serviceName string, instanceNumber int32) error {
	formattedServiceName := fmt.Sprintf("%s.instance.%d", serviceName, instanceNumber)

	// notify contract service to delete contracts
	ctx := context.Background()
	c.services.GetContractService().DeleteContract(ctx, formattedServiceName)

	return c.contracts.DeleteService(formattedServiceName)
}
