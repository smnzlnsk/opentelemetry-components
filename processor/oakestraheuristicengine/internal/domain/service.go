package domain

import (
	"context"

	"go.opentelemetry.io/collector/pdata/pmetric"
)

type MonitoringService interface {
	SaveMetrics(ctx context.Context, md pmetric.Metrics) error
	GetHostInstanceMetrics(ctx context.Context, host string, serviceInstance string) (DBHostMetrics, error)
}
