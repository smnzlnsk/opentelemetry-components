package internal

import (
	"context"

	pb "github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/proto"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
)

type ProcessorFactory interface {
	CreateDefaultConfig() Config
	CreateMetricsProcessor(
		ctx context.Context,
		settings processor.Settings,
		cfg Config) (MetricProcessor, error)
}

type MetricProcessor interface {
	Start(context.Context, component.Host) error
	ProcessMetrics(pmetric.Metrics) error
	Shutdown(context.Context) error
	RegisterService(serviceName string, instanceNumber int32, resource *pb.ResourceInfo) error
	DeleteService(serviceName string, instanceNumber int32) error
}
