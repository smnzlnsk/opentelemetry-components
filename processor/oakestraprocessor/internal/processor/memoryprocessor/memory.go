package memoryprocessor

import (
	"context"
	"fmt"
	"time"

	pb "github.com/smnzlnsk/monitoring-proto-lib/gen/go/monitoring_proto_lib/monitoring/v1"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/processor/memoryprocessor/internal/metadata"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/service"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

type MemoryMetricProcessor struct {
	contracts *internal.ContractState // create a per-service map of calculation contracts
	config    *Config
	logger    *zap.Logger
	cancel    context.CancelFunc
	settings  processor.Settings
	mb        *metadata.MetricsBuilder
	services  *service.Services
}

var _ internal.MetricProcessor = (*MemoryMetricProcessor)(nil)

// Define memory metrics as constants
const (
	memoryFormulaExpression = "([container.memory.usage] / [system.memory.usage]) * 1000000"
)

// Define required memory metric states
var requiredMemoryMetricStates = map[string]bool{
	"slab_reclaimable":   true,
	"slab_unreclaimable": true,
	"used":               true,
}

func (c *MemoryMetricProcessor) ProcessMetrics(metrics pmetric.Metrics) error {

	m, err := c.processMetrics(metrics)
	if err != nil {
		return err
	}

	m.ResourceMetrics().MoveAndAppendTo(metrics.ResourceMetrics())
	return nil
}

func (c *MemoryMetricProcessor) processMetrics(metrics pmetric.Metrics) (pmetric.Metrics, error) {
	// setup new calculation mechanism
	err := c.contracts.PopulateData(metrics)
	if err != nil {
		return metrics, err
	}

	results := c.contracts.Evaluate()

	for key, value := range results {
		rb := c.mb.NewResourceBuilder()
		rb.SetServiceName(key.Service)

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

	// initialize default contracts
	if err := c.contracts.GenerateDefaultContract(memoryFormulaExpression, requiredMemoryMetricStates); err != nil {
		return err
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
	services *service.Services,
) (internal.MetricProcessor, error) {
	return &MemoryMetricProcessor{
		contracts: internal.NewContractState(),
		config:    cfg.(*Config),
		settings:  set,
		logger:    set.Logger,
		services:  services,
	}, nil
}

func (c *MemoryMetricProcessor) RegisterService(serviceName string, instanceNumber int32, resource *pb.ResourceInfo, _ []*pb.CalculationRequest) error {
	// register default services in internal contract state
	err := c.contracts.RegisterService(fmt.Sprintf("%s.instance.%d", serviceName, instanceNumber), map[string]domain.CalculationContract{}, resource.Memory)
	if err != nil {
		return err
	}

	defContracts := c.contracts.GetDefaultContracts()
	contractsArray := make([]domain.CalculationContract, 0, len(defContracts))
	// Change service name from default to serviceName
	for _, contract := range defContracts {
		contract.Service = fmt.Sprintf("%s.instance.%d", serviceName, instanceNumber)
		contractsArray = append(contractsArray, contract)
	}

	// notify contract service to create contracts
	ctx := context.Background()
	err = c.services.ContractService.CreateMany(ctx, contractsArray)
	if err != nil {
		return err
	}

	return nil
}

func (c *MemoryMetricProcessor) DeleteService(serviceName string, instanceNumber int32) error {
	formattedServiceName := fmt.Sprintf("%s.instance.%d", serviceName, instanceNumber)

	// notify contract service to delete contracts
	ctx := context.Background()
	c.services.ContractService.DeleteContract(ctx, formattedServiceName)

	return c.contracts.DeleteService(formattedServiceName)
}
