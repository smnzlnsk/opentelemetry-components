package domain

import (
	"github.com/smnzlnsk/opentelemetry-components/internal/shared/metric"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type ProcessorStore interface {
	Add(processor Processor) error
	Get(identifier string) Processor
	GetAll() map[string]Processor
}

type MetricStore interface {
	Store(key metric.Key, value float64)
	Save(md pmetric.Metrics) error
	GetValueForMetricKey(key metric.Key) float64
	GetValueMapByMetricKey() map[metric.Key]float64
	GetValueMapByString() map[string]interface{}
}

type MemoryStore interface {
	GetNodeStore(nodeID string) NodeStore
}

type NodeStore interface {
}
