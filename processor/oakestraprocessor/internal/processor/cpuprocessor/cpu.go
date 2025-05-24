package cpuprocessor

import (
	"context"
	"fmt"
	"time"

	pb "github.com/smnzlnsk/monitoring-proto-lib/gen/go/monitoring_proto_lib/monitoring/v1"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/domain"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/processor/cpuprocessor/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

type CPUMetricProcessor struct {
	contracts *domain.ContractState // create a per-service map of calculation contracts
	config    *Config
	logger    *zap.Logger
	cancel    context.CancelFunc
	settings  processor.Settings
	mb        *metadata.MetricsBuilder
	services  domain.Services
}

var _ internal.MetricProcessor = (*CPUMetricProcessor)(nil)

func (c *CPUMetricProcessor) ProcessMetrics(metrics pmetric.Metrics) error {

	m, err := c.processMetrics(metrics)
	if err != nil {
		return err
	}

	m.ResourceMetrics().MoveAndAppendTo(metrics.ResourceMetrics())
	return nil
}

func (c *CPUMetricProcessor) processMetrics(_ pmetric.Metrics) (pmetric.Metrics, error) {
	// setup new calculation mechanism

	results := c.contracts.Evaluate()

	for key, value := range results {
		rb := c.mb.NewResourceBuilder()
		rb.SetServiceName(key.Service)
		rb.SetContainerID(key.Service)
		c.mb.RecordServiceCPUUtilisationDataPoint(
			pcommon.NewTimestampFromTime(time.Now()),
			value,
			metadata.MapAttributeState[key.State],
		)

		// set resources
		c.mb.EmitForResource(metadata.WithResource(rb.Emit()))
	}
	return c.mb.Emit(), nil
}

func (c *CPUMetricProcessor) Shutdown(_ context.Context) error {
	if c.cancel != nil {
		c.cancel()
	}
	c.logger.Info("Stopped CPU Processor")
	return nil
}

func (c *CPUMetricProcessor) Start(ctx context.Context, _ component.Host) error {
	_, c.cancel = context.WithCancel(ctx)

	// sync contracts
	c.contracts.Sync()

	defaultContracts := []struct {
		Formula string
		States  []string
	}{
		{
			Formula: "((([container.cpu.time(0)] - [container.cpu.time(1)]) / 1000000000) / ([system.cpu.time(0)] - [system.cpu.time(1)])) * 100",
			States:  []string{"user", "system"},
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
	c.logger.Info("Started CPU Processor")
	return nil
}

func newCPUMetricProcessor(
	_ context.Context,
	set processor.Settings,
	cfg internal.Config,
	services domain.Services,
	dm domain.DatapointManager,
) (internal.MetricProcessor, error) {

	return &CPUMetricProcessor{
		contracts: domain.NewContractState(TypeStr, set.Logger, services, dm),
		config:    cfg.(*Config),
		settings:  set,
		logger:    set.Logger,
		services:  services,
	}, nil
}

func (c *CPUMetricProcessor) RegisterService(serviceName string, instanceNumber int32, resource *pb.ResourceInfo, _ []*pb.CalculationRequest) error {
	// register service in internal contract state
	err := c.contracts.RegisterService(fmt.Sprintf("%s.instance.%d", serviceName, instanceNumber), []domain.CalculationContract{}, resource.Cpu)
	if err != nil {
		return err
	}

	// Register default contracts with contract service
	defContracts := c.contracts.GetDefaultContracts()
	contractsArray := make([]domain.CalculationContract, 0, len(defContracts))
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

func (c *CPUMetricProcessor) DeleteService(serviceName string, instanceNumber int32) error {
	formattedServiceName := fmt.Sprintf("%s.instance.%d", serviceName, instanceNumber)

	// notify contract service to delete contracts
	ctx := context.Background()
	c.services.GetContractService().DeleteContract(ctx, formattedServiceName)

	return c.contracts.DeleteService(formattedServiceName)
}
