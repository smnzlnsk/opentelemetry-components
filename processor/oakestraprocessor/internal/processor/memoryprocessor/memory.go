package memoryprocessor

import (
	"context"
	"fmt"
	"time"

	pb "github.com/smnzlnsk/monitoring-proto-lib/gen/go/monitoring_proto_lib/monitoring/v1"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/processor/memoryprocessor/internal/metadata"
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
) (internal.MetricProcessor, error) {
	return &MemoryMetricProcessor{
		contracts: internal.NewContractState(),
		config:    cfg.(*Config),
		settings:  set,
		logger:    set.Logger,
	}, nil
}

func (c *MemoryMetricProcessor) RegisterService(serviceName string, instanceNumber int32, resource *pb.ResourceInfo, _ []*pb.CalculationRequest) error {
	return c.contracts.RegisterService(fmt.Sprintf("%s.instance.%d", serviceName, instanceNumber), map[string]internal.CalculationContract{}, resource.Memory)
}

func (c *MemoryMetricProcessor) DeleteService(serviceName string, instanceNumber int32) error {
	return c.contracts.DeleteService(fmt.Sprintf("%s.instance.%d", serviceName, instanceNumber))
}
