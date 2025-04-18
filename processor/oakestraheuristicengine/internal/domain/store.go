package domain

import (
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type ProcessorStore interface {
	Add(processor Processor) error
	Get(identifier string) Processor
	GetAll() map[string]Processor
}

type MetricStore interface {
	Store(key MetricKey, value float64)
	Save(md pmetric.Metrics) error
	GetValueForMetricKey(key MetricKey) float64
	GetValueMapByMetricKey() map[MetricKey]float64
	GetValueMapByString() map[string]interface{}
}

type MemoryStore interface {
	GetNodeStore(nodeID string) NodeStore
}

type NodeStore interface {
}
