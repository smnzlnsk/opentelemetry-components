package interfaces

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type ProcessorStore interface {
	Add(processor Processor) error
	Get(identifier string) Processor
	GetAll() map[string]Processor
}

type MetricStore interface {
	Store(key types.MetricKey, value float64)
	Save(md pmetric.Metrics) error
	GetValueForMetricKey(key types.MetricKey) float64
	GetValueMapByMetricKey() map[types.MetricKey]float64
	GetValueMapByString() map[string]interface{}
}
