package memoryprocessor

import (
	"context"
	"fmt"
	"time"

	pb "github.com/smnzlnsk/monitoring-proto-lib/gen/go/monitoring_proto_lib/monitoring/v1"
	"github.com/smnzlnsk/opentelemetry-components/pkg/calculation"
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

	if len(results) == 0 {
		return c.mb.Emit(), nil
	}

	// Group results by service name
	serviceGroups := make(map[string][]struct {
		Key   calculation.ResultKey
		Value float64
	})

	for key, value := range results {
		serviceName := key.Service
		if serviceGroups[serviceName] == nil {
			serviceGroups[serviceName] = make([]struct {
				Key   calculation.ResultKey
				Value float64
			}, 0)
		}
		serviceGroups[serviceName] = append(serviceGroups[serviceName], struct {
			Key   calculation.ResultKey
			Value float64
		}{Key: key, Value: value})
	}

	// Create a final metrics object to hold all resource metrics
	finalMetrics := pmetric.NewMetrics()

	// Create a separate resource metric for each unique service
	for serviceName, serviceData := range serviceGroups {
		// Create a new metrics builder for this service
		mb := metadata.NewMetricsBuilder(c.config.MetricsBuilderConfig, receiver.Settings{TelemetrySettings: c.settings.TelemetrySettings})

		// Add datapoints for this service
		for _, data := range serviceData {
			key := data.Key
			value := data.Value
			mb.RecordServiceMemoryUtilisationDataPoint(
				pcommon.NewTimestampFromTime(time.Now()),
				value,
				metadata.MapAttributeState[key.State],
			)
		}

		// Create resource for this service
		rb := mb.NewResourceBuilder()
		rb.SetServiceName(serviceName)
		rb.SetContainerID(serviceName)

		// Emit metrics for this service and add to final metrics
		serviceMetrics := mb.Emit(metadata.WithResource(rb.Emit()))
		serviceMetrics.ResourceMetrics().MoveAndAppendTo(finalMetrics.ResourceMetrics())
	}

	return finalMetrics, nil
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
			Formula: "([container.memory.usage] / [system.memory.limit{default}]) * 100",
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
