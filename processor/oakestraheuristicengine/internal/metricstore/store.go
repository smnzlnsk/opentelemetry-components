package metricstore

import (
	"container/list"
	"fmt"
	"math"
	"sync"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/constants"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type MetricStats struct {
	Average float64
	StdDev  float64
	Sum     float64
	Count   int
	Min     float64
	Max     float64
}

// MetricStore provides thread-safe storage and management of metrics
type metricStore struct {
	mu      sync.RWMutex
	metrics map[types.MetricKey]*list.List
	logger  *zap.Logger
}

// NewMetricStore creates a new MetricStore instance
func NewMetricStore(logger *zap.Logger) interfaces.MetricStore {
	return &metricStore{
		metrics: make(map[types.MetricKey]*list.List),
		logger:  logger,
	}
}

// Store adds a metric to the store with the current timestamp
func (ms *metricStore) Store(key types.MetricKey, value float64) {
	if _, exists := ms.metrics[key]; !exists {
		ms.metrics[key] = list.New()
	}
	ms.metrics[key].PushBack(value)
	if ms.metrics[key].Len() > 5 {
		ms.metrics[key].Remove(ms.metrics[key].Front())
	}

	if key.Type == constants.MetricValueTypeRaw {
		ms.calculateStats(key, value)
	}
}

func (ms *metricStore) calculateStats(key types.MetricKey, value float64) {
	values := ms.metrics[key]

	// Calculate average
	var sum float64
	min, max := values.Front().Value.(float64), values.Front().Value.(float64)
	for e := values.Front(); e != nil; e = e.Next() {
		v := e.Value.(float64)
		sum += v
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	avg := sum / float64(values.Len())

	// Calculate standard deviation
	var sumSquareDiff float64
	for e := values.Front(); e != nil; e = e.Next() {
		v := e.Value.(float64)
		diff := v - avg
		sumSquareDiff += diff * diff
	}
	stddev := math.Sqrt(sumSquareDiff / float64(values.Len()))

	ms.Store(types.MetricKey{Name: key.Name, State: key.State, Type: constants.MetricValueTypeAvg}, avg)
	ms.Store(types.MetricKey{Name: key.Name, State: key.State, Type: constants.MetricValueTypeStdDev}, stddev)
	ms.Store(types.MetricKey{Name: key.Name, State: key.State, Type: constants.MetricValueTypeCount}, float64(values.Len()))
	ms.Store(types.MetricKey{Name: key.Name, State: key.State, Type: constants.MetricValueTypeMin}, min)
	ms.Store(types.MetricKey{Name: key.Name, State: key.State, Type: constants.MetricValueTypeMax}, max)
}

func (ms *metricStore) Save(md pmetric.Metrics) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Iterate through all resource metrics
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)

		// Iterate through all scope metrics
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)

			// Iterate through all metrics
			for k := 0; k < sm.Metrics().Len(); k++ {
				metric := sm.Metrics().At(k)
				err := ms.extractDataPoints(metric)
				if err != nil {
					ms.logger.Error("failed to extract data points", zap.Error(err))
				}
			}
		}
	}

	return nil
}

// Get retrieves metrics for a given key
func (ms *metricStore) GetValueForMetricKey(key types.MetricKey) float64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if metrics, exists := ms.metrics[key]; exists {
		if metrics.Len() > 0 {
			return metrics.Back().Value.(float64)
		}
	}
	return -1
}

// extractDataPoints extracts values from different metric types
func (ms *metricStore) extractDataPoints(metric pmetric.Metric) error {
	name := metric.Name()

	var value float64

	switch metric.Type() {
	case pmetric.MetricTypeGauge:
		dps := metric.Gauge().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			dp := dps.At(i)
			state, _ := dp.Attributes().Get("state")
			switch dp.ValueType() {
			case pmetric.NumberDataPointValueTypeDouble:
				value = dp.DoubleValue()
			case pmetric.NumberDataPointValueTypeInt:
				value = float64(dp.IntValue())
			}
			ms.Store(types.MetricKey{Name: name, State: state.Str(), Type: constants.MetricValueTypeRaw}, value)
		}

	case pmetric.MetricTypeSum:
		dps := metric.Sum().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			dp := dps.At(i)
			state, _ := dp.Attributes().Get("state")
			switch dp.ValueType() {
			case pmetric.NumberDataPointValueTypeDouble:
				value = dp.DoubleValue()
			case pmetric.NumberDataPointValueTypeInt:
				value = float64(dp.IntValue())
			}
			ms.Store(types.MetricKey{Name: name, State: state.Str(), Type: constants.MetricValueTypeRaw}, value)
		}

	case pmetric.MetricTypeHistogram:
		dps := metric.Histogram().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			dp := dps.At(i)
			state, _ := dp.Attributes().Get("state")
			// For histograms, we'll use the sum
			value = dp.Sum()
			ms.Store(types.MetricKey{Name: name, State: state.Str(), Type: constants.MetricValueTypeRaw}, value)
		}

	case pmetric.MetricTypeSummary:
		dps := metric.Summary().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			dp := dps.At(i)
			state, _ := dp.Attributes().Get("state")
			// For summaries, we'll use the sum
			value = dp.Sum()
			ms.Store(types.MetricKey{Name: name, State: state.Str(), Type: constants.MetricValueTypeRaw}, value)
		}
	}

	return nil
}

func (ms *metricStore) GetValueMapByMetricKey() map[types.MetricKey]float64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return nil
}

func (ms *metricStore) GetValueMapByString() map[string]interface{} {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	stateString := ""

	values := make(map[string]interface{})
	for key, list := range ms.metrics {

		if key.State != "" {
			stateString = fmt.Sprintf("|%s", key.State)
		}
		values[fmt.Sprintf("%s|%s%s", key.Name, key.Type, stateString)] = list.Back().Value.(float64)
	}
	return values
}
