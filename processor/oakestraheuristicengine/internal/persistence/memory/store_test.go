package memory

import (
	"testing"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

func TestMetricStore_Basic(t *testing.T) {
	// Create a test logger
	logger := zap.NewNop()

	// Create a new metric store
	ms := NewMetricStore(logger)

	// Test storing and retrieving a simple metric
	key := domain.MetricKey{
		Name:  "test_metric",
		State: "normal",
		Type:  domain.MetricValueTypeRaw,
	}

	// Store a value
	ms.Store(key, 42.0)

	// Retrieve the value
	value := ms.GetValueForMetricKey(key)

	// Assert that the retrieved value matches what we stored
	assert.Equal(t, 42.0, value)

	// Test that derived metrics were calculated
	avgKey := domain.MetricKey{
		Name:  "test_metric",
		State: "normal",
		Type:  domain.MetricValueTypeAvg,
	}

	avgValue := ms.GetValueForMetricKey(avgKey)
	assert.Equal(t, 42.0, avgValue)
}

func TestMetricStore_Complex(t *testing.T) {
	// Create a test logger
	logger := zap.NewNop()

	// Create a new metric store
	ms := NewMetricStore(logger)

	// Test storing multiple values and checking statistics
	key := domain.MetricKey{
		Name:  "complex_metric",
		State: "warning",
		Type:  domain.MetricValueTypeRaw,
	}

	// Store multiple values to test statistics calculation
	testValues := []float64{10.0, 20.0, 30.0, 40.0, 50.0}
	for _, val := range testValues {
		ms.Store(key, val)
	}

	// Test raw value (should be the last value stored)
	rawValue := ms.GetValueForMetricKey(key)
	assert.Equal(t, 50.0, rawValue)

	// Test average
	avgKey := domain.MetricKey{
		Name:  "complex_metric",
		State: "warning",
		Type:  domain.MetricValueTypeAvg,
	}
	avgValue := ms.GetValueForMetricKey(avgKey)
	assert.Equal(t, 30.0, avgValue)

	// Test min
	minKey := domain.MetricKey{
		Name:  "complex_metric",
		State: "warning",
		Type:  domain.MetricValueTypeMin,
	}
	minValue := ms.GetValueForMetricKey(minKey)
	assert.Equal(t, 10.0, minValue)

	// Test max
	maxKey := domain.MetricKey{
		Name:  "complex_metric",
		State: "warning",
		Type:  domain.MetricValueTypeMax,
	}
	maxValue := ms.GetValueForMetricKey(maxKey)
	assert.Equal(t, 50.0, maxValue)

	// Test count
	countKey := domain.MetricKey{
		Name:  "complex_metric",
		State: "warning",
		Type:  domain.MetricValueTypeCount,
	}
	countValue := ms.GetValueForMetricKey(countKey)
	assert.Equal(t, 5.0, countValue)

	// Test stddev
	stddevKey := domain.MetricKey{
		Name:  "complex_metric",
		State: "warning",
		Type:  domain.MetricValueTypeStdDev,
	}
	stddevValue := ms.GetValueForMetricKey(stddevKey)
	// Expected stddev for [10,20,30,40,50] with mean 30 is sqrt(200) = 14.142...
	assert.InDelta(t, 14.142, stddevValue, 0.001)

	// Test GetValueMapByString
	valueMap := ms.GetValueMapByString()
	assert.NotNil(t, valueMap)

	// Check that our metrics are in the map with correct format
	assert.Contains(t, valueMap, "complex_metric|raw|warning")
	assert.Equal(t, 50.0, valueMap["complex_metric|raw|warning"])
}

func TestMetricStore_Save(t *testing.T) {
	// Create a test logger
	logger := zap.NewNop()

	// Create a new metric store
	ms := NewMetricStore(logger)

	// Create test metrics
	metrics := createTestMetrics()

	// Save metrics
	err := ms.Save(metrics)
	require.NoError(t, err)

	// Verify that metrics were stored correctly
	gaugeKey := domain.MetricKey{
		Name:  "test_gauge",
		State: "normal",
		Type:  domain.MetricValueTypeRaw,
	}
	gaugeValue := ms.GetValueForMetricKey(gaugeKey)
	assert.Equal(t, 123.45, gaugeValue)

	sumKey := domain.MetricKey{
		Name:  "test_sum",
		State: "critical",
		Type:  domain.MetricValueTypeRaw,
	}
	sumValue := ms.GetValueForMetricKey(sumKey)
	assert.Equal(t, 678.0, sumValue)
}

func TestMetricStore_Error(t *testing.T) {
	// Create a test logger
	logger := zap.NewNop()

	// Create a new metric store
	ms := NewMetricStore(logger)

	// Test retrieving a non-existent metric
	nonExistentKey := domain.MetricKey{
		Name:  "non_existent",
		State: "unknown",
		Type:  domain.MetricValueTypeRaw,
	}

	value := ms.GetValueForMetricKey(nonExistentKey)
	assert.Equal(t, -1.0, value, "Non-existent metrics should return -1")

	// Test with invalid metrics data
	invalidMetrics := pmetric.NewMetrics()
	rm := invalidMetrics.ResourceMetrics().AppendEmpty()
	sm := rm.ScopeMetrics().AppendEmpty()
	metric := sm.Metrics().AppendEmpty()

	// Set up an invalid metric (missing required fields)
	metric.SetName("invalid_metric")
	metric.SetEmptyGauge()

	// This shouldn't cause a panic, but the metric won't be stored properly
	err := ms.Save(invalidMetrics)
	assert.NoError(t, err, "Save should not return an error even with invalid metrics")

	// The invalid metric should not be stored
	invalidKey := domain.MetricKey{
		Name:  "invalid_metric",
		State: "",
		Type:  domain.MetricValueTypeRaw,
	}

	invalidValue := ms.GetValueForMetricKey(invalidKey)
	assert.Equal(t, -1.0, invalidValue, "Invalid metrics should not be stored")
}

// Helper function to create test metrics
func createTestMetrics() pmetric.Metrics {
	metrics := pmetric.NewMetrics()

	// Add a resource
	rm := metrics.ResourceMetrics().AppendEmpty()

	// Add a scope
	sm := rm.ScopeMetrics().AppendEmpty()

	// Add a gauge metric
	gaugeMetric := sm.Metrics().AppendEmpty()
	gaugeMetric.SetName("test_gauge")
	gauge := gaugeMetric.SetEmptyGauge()
	gaugeDP := gauge.DataPoints().AppendEmpty()
	gaugeDP.SetDoubleValue(123.45)
	gaugeDP.Attributes().PutStr("state", "normal")

	// Add a sum metric
	sumMetric := sm.Metrics().AppendEmpty()
	sumMetric.SetName("test_sum")
	sum := sumMetric.SetEmptySum()
	sumDP := sum.DataPoints().AppendEmpty()
	sumDP.SetIntValue(678)
	sumDP.Attributes().PutStr("state", "critical")

	// Add a histogram metric
	histMetric := sm.Metrics().AppendEmpty()
	histMetric.SetName("test_histogram")
	hist := histMetric.SetEmptyHistogram()
	histDP := hist.DataPoints().AppendEmpty()
	histDP.SetSum(1000.0)
	histDP.Attributes().PutStr("state", "warning")

	// Add a summary metric
	summaryMetric := sm.Metrics().AppendEmpty()
	summaryMetric.SetName("test_summary")
	summary := summaryMetric.SetEmptySummary()
	summaryDP := summary.DataPoints().AppendEmpty()
	summaryDP.SetSum(500.0)
	summaryDP.Attributes().PutStr("state", "unknown")

	return metrics
}

func TestMetricStore_GetValueMapByMetricKey(t *testing.T) {
	// Create a test logger
	logger := zap.NewNop()

	// Create a new metric store
	ms := NewMetricStore(logger)

	// Store some test metrics
	ms.Store(domain.MetricKey{Name: "test1", State: "normal", Type: domain.MetricValueTypeRaw}, 100.0)
	ms.Store(domain.MetricKey{Name: "test2", State: "warning", Type: domain.MetricValueTypeRaw}, 200.0)

	// The current implementation returns nil, so we should test that
	valueMap := ms.GetValueMapByMetricKey()
	assert.Nil(t, valueMap, "Current implementation returns nil")

	// This test can be expanded when the implementation is completed
}

func TestMetricStore_HistoryLimit(t *testing.T) {
	// Create a test logger
	logger := zap.NewNop()

	// Create a new metric store
	ms := NewMetricStore(logger)

	// Test that the store only keeps the last 5 values
	key := domain.MetricKey{
		Name:  "history_test",
		State: "normal",
		Type:  domain.MetricValueTypeRaw,
	}

	// Store more than 5 values
	for i := 0; i < 10; i++ {
		ms.Store(key, float64(i))
	}

	// The last value should be 9
	value := ms.GetValueForMetricKey(key)
	assert.Equal(t, 9.0, value)

	// The average should be calculated from the last 5 values (5,6,7,8,9)
	avgKey := domain.MetricKey{
		Name:  "history_test",
		State: "normal",
		Type:  domain.MetricValueTypeAvg,
	}
	avgValue := ms.GetValueForMetricKey(avgKey)
	assert.Equal(t, 7.0, avgValue)
}

func TestMetricStore_ExtractDataPoints(t *testing.T) {
	// Create a test logger
	logger := zap.NewNop()

	// Create a new metric store
	ms := NewMetricStore(logger).(*metricStore)

	// Create a test metric
	metric := pmetric.NewMetric()
	metric.SetName("test_extract")
	gauge := metric.SetEmptyGauge()
	dp := gauge.DataPoints().AppendEmpty()
	dp.SetDoubleValue(42.0)
	dp.Attributes().PutStr("state", "normal")

	// Extract data points
	err := ms.extractDataPoints(metric)
	assert.NoError(t, err)

	// Verify the metric was stored
	key := domain.MetricKey{
		Name:  "test_extract",
		State: "normal",
		Type:  domain.MetricValueTypeRaw,
	}
	value := ms.GetValueForMetricKey(key)
	assert.Equal(t, 42.0, value)
}
