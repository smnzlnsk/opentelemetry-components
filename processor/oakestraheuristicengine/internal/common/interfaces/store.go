package interfaces

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type TreeStore interface {
	Add(identifier string, decisionTree DecisionTree) error
	Get(identifier string) DecisionTree
}

type MetricStore interface {
	Store(key types.MetricKey, value float64)
	Save(md pmetric.Metrics) error
	GetValueForMetricKey(key types.MetricKey) float64
	GetValueMapByMetricKey() map[types.MetricKey]float64
	GetValueMapByString() map[string]interface{}
}
