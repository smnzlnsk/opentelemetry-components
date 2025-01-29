package applicationprocessor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormulaToMetricMap(t *testing.T) {
	t.Run("new map should be empty", func(t *testing.T) {
		ftm := NewFormulaToMetricMap()
		result := ftm.GetMetricName("service1", "formula1")
		assert.Empty(t, result.MetricName)
		assert.Empty(t, result.MetricUnit)
	})

	t.Run("should store and retrieve metric metadata", func(t *testing.T) {
		ftm := NewFormulaToMetricMap()

		// Add a metric
		ftm.AddMetric("service1", "formula1", "metric1", "bytes")

		// Retrieve the metric
		result := ftm.GetMetricName("service1", "formula1")
		assert.Equal(t, "metric1", result.MetricName)
		assert.Equal(t, "bytes", result.MetricUnit)
	})

	t.Run("should handle multiple entries", func(t *testing.T) {
		ftm := NewFormulaToMetricMap()

		// Add multiple metrics
		ftm.AddMetric("service1", "formula1", "metric1", "bytes")
		ftm.AddMetric("service2", "formula2", "metric2", "seconds")
		ftm.AddMetric("service1", "formula2", "metric3", "count")

		// Verify each entry
		result1 := ftm.GetMetricName("service1", "formula1")
		assert.Equal(t, "metric1", result1.MetricName)
		assert.Equal(t, "bytes", result1.MetricUnit)

		result2 := ftm.GetMetricName("service2", "formula2")
		assert.Equal(t, "metric2", result2.MetricName)
		assert.Equal(t, "seconds", result2.MetricUnit)

		result3 := ftm.GetMetricName("service1", "formula2")
		assert.Equal(t, "metric3", result3.MetricName)
		assert.Equal(t, "count", result3.MetricUnit)
	})

	t.Run("should handle overwriting existing entries", func(t *testing.T) {
		ftm := NewFormulaToMetricMap()

		// Add initial metric
		ftm.AddMetric("service1", "formula1", "metric1", "bytes")

		// Overwrite with new values
		ftm.AddMetric("service1", "formula1", "metric2", "seconds")

		// Verify the new values
		result := ftm.GetMetricName("service1", "formula1")
		assert.Equal(t, "metric2", result.MetricName)
		assert.Equal(t, "seconds", result.MetricUnit)
	})

	t.Run("should return empty metadata for non-existent entries", func(t *testing.T) {
		ftm := NewFormulaToMetricMap()

		ftm.AddMetric("service1", "formula1", "metric1", "bytes")

		// Try to get non-existent entry
		result := ftm.GetMetricName("service2", "formula1")
		assert.Empty(t, result.MetricName)
		assert.Empty(t, result.MetricUnit)
	})

	t.Run("should delete all metrics for a service", func(t *testing.T) {
		ftm := NewFormulaToMetricMap()

		// Add multiple metrics for the same service
		ftm.AddMetric("service1", "formula1", "metric1", "bytes")
		ftm.AddMetric("service1", "formula2", "metric2", "seconds")

		// Delete all metrics for service1
		ftm.DeleteMetric("service1")

		// Verify that all metrics for service1 are deleted
		result1 := ftm.GetMetricName("service1", "formula1")
		assert.Empty(t, result1.MetricName)
		assert.Empty(t, result1.MetricUnit)

		result2 := ftm.GetMetricName("service1", "formula2")
		assert.Empty(t, result2.MetricName)
		assert.Empty(t, result2.MetricUnit)
	})

	t.Run("should handle delete non-existent service gracefully", func(t *testing.T) {
		ftm := NewFormulaToMetricMap()

		// Attempt to delete metrics for a non-existent service
		ftm.DeleteMetric("service1")

		// Verify that no error occurs and the state remains unchanged
		result1 := ftm.GetMetricName("service1", "formula1")
		assert.Empty(t, result1.MetricName)
		assert.Empty(t, result1.MetricUnit)
	})
}
