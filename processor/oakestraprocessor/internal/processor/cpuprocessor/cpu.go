package cpuprocessor

import (
	"context"
	"fmt"
	"time"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/processor/cpuprocessor/internal/metadata"
	pb "github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/proto"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

type CPUMetricProcessor struct {
	contracts *internal.ContractState // create a per-service map of calculation contracts
	config    *Config
	logger    *zap.Logger
	cancel    context.CancelFunc
	settings  processor.Settings
	mb        *metadata.MetricsBuilder
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

func (c *CPUMetricProcessor) processMetrics(metrics pmetric.Metrics) (pmetric.Metrics, error) {
	// setup new calculation mechanism
	err := c.contracts.PopulateData(metrics)
	if err != nil {
		return metrics, err
	}

	results := c.contracts.Evaluate()

	for service, f := range results {
		rb := c.mb.NewResourceBuilder()
		rb.SetServiceName(service)
		for _, state := range f {
			for s, v := range state {
				c.mb.RecordServiceCPUUtilisationDataPoint(
					pcommon.NewTimestampFromTime(time.Now()),
					v,
					metadata.MapAttributeState[s],
				)
			}
		}
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
	ctx, c.cancel = context.WithCancel(ctx)

	// initialize default contracts
	if err := c.contracts.GenerateDefaultContract(
		"[container.cpu.time] / [system.cpu.time]",
		map[string]bool{
			"user":   true,
			"system": true},
	); err != nil {
		return err
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
) (internal.MetricProcessor, error) {

	/*
		metricFilter := map[string]bool{
			"container.cpu.time":       true,
			"system.cpu.time":          true,
			"system.cpu.logical.count": true,
			"system.cpu.utilization":   false,
		}
		stateFilter := map[string]bool{
			"system": true,
			"user":   true,
		}
	*/

	return &CPUMetricProcessor{
		contracts: internal.NewContractState(),
		config:    cfg.(*Config),
		settings:  set,
		logger:    set.Logger,
	}, nil
}

func (c *CPUMetricProcessor) RegisterService(serviceName string, instanceNumber int32, resource *pb.ResourceInfo) error {
	return c.contracts.RegisterService(fmt.Sprintf("%s.instance.%d", serviceName, instanceNumber), nil)
}

func (c *CPUMetricProcessor) DeleteService(serviceName string, instanceNumber int32) error {
	return c.contracts.DeleteService(fmt.Sprintf("%s.instance.%d", serviceName, instanceNumber))
}
