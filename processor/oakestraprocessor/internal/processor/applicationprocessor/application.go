package applicationprocessor

import (
	"context"
	"fmt"

	pb "github.com/smnzlnsk/monitoring-proto-lib/gen/go/monitoring_proto_lib/monitoring/v1"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/processor/applicationprocessor/internal/builder"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
)

type ApplicationMetricProcessor struct {
	contracts          *internal.ContractState // create a per-service map of calculation contracts
	formulaToMetricMap *FormulaToMetricMap
	config             *Config
	logger             *zap.Logger
	cancel             context.CancelFunc
	mb                 *builder.MetricsBuilder
}

var _ internal.MetricProcessor = (*ApplicationMetricProcessor)(nil)

func (c *ApplicationMetricProcessor) ProcessMetrics(metrics pmetric.Metrics) error {
	m, err := c.processMetrics(metrics)
	if err != nil {
		return err
	}
	m.ResourceMetrics().MoveAndAppendTo(metrics.ResourceMetrics())
	return nil
}

func (c *ApplicationMetricProcessor) processMetrics(metrics pmetric.Metrics) (pmetric.Metrics, error) {
	// setup new calculation mechanism
	err := c.contracts.PopulateData(metrics)
	if err != nil {
		return metrics, err
	}

	results := c.contracts.Evaluate()

	for key, value := range results {
		mmetadata := c.formulaToMetricMap.GetMetricName(key.Service, key.Formula)
		// Create new resource and metric builders for each service
		rb := c.mb.NewResourceBuilder()
		mbb := rb.NewMetricBuilder()

		rb.SetServiceName(key.Service)
		mbb.AddSum(
			mmetadata.MetricName,
			key.State,
			"Calculated metric for "+key.Formula,
			mmetadata.MetricUnit,
			value,
		)
	}

	return c.mb.Emit(), nil
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
		contracts:          internal.NewContractState(),
		formulaToMetricMap: NewFormulaToMetricMap(),
		config:             cfg.(*Config),
		logger:             set.Logger,
		mb:                 builder.NewMetricsBuilder(),
	}, nil
}

func (c *ApplicationMetricProcessor) RegisterService(serviceName string, instanceNumber int32, resource *pb.ResourceInfo, calculationRequests []*pb.CalculationRequest) error {
	formattedServiceName := fmt.Sprintf("%s.instance.%d", serviceName, instanceNumber)
	contracts := internal.NewCalculationContractsFromProto(formattedServiceName, calculationRequests)

	for _, req := range calculationRequests {
		c.formulaToMetricMap.AddMetric(formattedServiceName, req.Formula, req.MetricName, req.Unit)
	}

	return c.contracts.RegisterService(formattedServiceName, contracts, "1") // INFO: no normalization is to be done for application processor, so we set it to 1 (for now)
}

func (c *ApplicationMetricProcessor) DeleteService(serviceName string, instanceNumber int32) error {
	formattedServiceName := fmt.Sprintf("%s.instance.%d", serviceName, instanceNumber)

	c.formulaToMetricMap.DeleteMetric(formattedServiceName)

	return c.contracts.DeleteService(formattedServiceName)
}
