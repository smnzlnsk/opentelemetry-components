package metricstore

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type MetricStore interface {
	Store(key string, metric pmetric.Metrics)
	Get(key string) []MetricEntry
	Export(ctx context.Context) error
	Cleanup(retention time.Duration)
}

// MetricStore provides thread-safe storage and management of metrics
type metricStore struct {
	mu      sync.RWMutex
	metrics map[string][]MetricEntry
	logger  *zap.Logger
	// TODO: add db config
	// dbConfig DBConfig
}

// MetricEntry represents a single metric package entry with timestamp
type MetricEntry struct {
	Metric    pmetric.Metrics
	Timestamp time.Time
}

// NewMetricStore creates a new MetricStore instance
func NewMetricStore(logger *zap.Logger) MetricStore {
	return &metricStore{
		metrics: make(map[string][]MetricEntry),
		logger:  logger,
	}
}

// Store adds a metric to the store with the current timestamp
func (ms *metricStore) Store(key string, metric pmetric.Metrics) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	entry := MetricEntry{
		Metric:    metric,
		Timestamp: time.Now(),
	}

	if _, exists := ms.metrics[key]; !exists {
		ms.metrics[key] = make([]MetricEntry, 0)
	}
	ms.metrics[key] = append(ms.metrics[key], entry)
}

// Get retrieves metrics for a given key
func (ms *metricStore) Get(key string) []MetricEntry {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if metrics, exists := ms.metrics[key]; exists {
		return metrics
	}
	return nil
}

// Export would handle exporting metrics to a database
// This is a placeholder that you can implement based on your specific database needs
func (ms *metricStore) Export(ctx context.Context) error {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	// TODO: Implement database export logic
	// For example:
	// 1. Connect to your database
	// 2. Format metrics for storage
	// 3. Batch insert metrics
	// 4. Handle errors and retries

	return nil
}

// Cleanup removes old metrics based on a retention period
func (ms *metricStore) Cleanup(retention time.Duration) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	cutoff := time.Now().Add(-retention)

	for key, entries := range ms.metrics {
		filtered := make([]MetricEntry, 0)
		for _, entry := range entries {
			if entry.Timestamp.After(cutoff) {
				filtered = append(filtered, entry)
			}
		}
		if len(filtered) > 0 {
			ms.metrics[key] = filtered
		} else {
			delete(ms.metrics, key)
		}
	}
}
