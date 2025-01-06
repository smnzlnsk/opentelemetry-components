package applicationprocessor

import (
	"context"
	"fmt"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
	pb "github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/proto"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
)

type ApplicationMetricProcessor struct {
	contracts *internal.ContractState // create a per-service map of calculation contracts
	config    *Config
	logger    *zap.Logger
	cancel    context.CancelFunc
}

var _ internal.MetricProcessor = (*ApplicationMetricProcessor)(nil)

func (c *ApplicationMetricProcessor) ProcessMetrics(metrics pmetric.Metrics) error {
	return nil
}

func (c *ApplicationMetricProcessor) Shutdown(ctx context.Context) error {
	if c.cancel != nil {
		c.cancel()
	}
	c.logger.Info("Stopped Application Processor")
	return nil
}

func (c *ApplicationMetricProcessor) Start(ctx context.Context, _ component.Host) error {
	_, c.cancel = context.WithCancel(ctx)
	c.logger.Info("Started Application Processor")
	return nil
}

func newApplicationMetricProcessor(
	_ context.Context,
	set processor.Settings,
	cfg internal.Config,
) (internal.MetricProcessor, error) {
	return &ApplicationMetricProcessor{
		contracts: internal.NewContractState(),
		config:    cfg.(*Config),
		logger:    set.Logger,
	}, nil
}

func (c *ApplicationMetricProcessor) RegisterService(serviceName string, instanceNumber int32, resource *pb.ResourceInfo) error {
	return c.contracts.RegisterService(fmt.Sprintf("%s.instance.%d", serviceName, instanceNumber), nil)
}

func (c *ApplicationMetricProcessor) DeleteService(serviceName string, instanceNumber int32) error {
	return nil
}
