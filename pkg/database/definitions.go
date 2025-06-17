package database

import (
	"fmt"
	"sync"
	"time"

	"github.com/smnzlnsk/opentelemetry-components/pkg/metric"
)

type MetricDatapoint struct {
	Value     float64   `json:"value" bson:"value"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

type MetricDatapoints struct {
	Identifier metric.Key        `json:"identifier" bson:"identifier"`
	Datapoints []MetricDatapoint `json:"datapoints" bson:"datapoints"`
}

type ServiceInstanceMetrics struct {
	JobName        string             `json:"job_name" bson:"job_name"`
	InstanceNumber int                `json:"instance_number" bson:"instance_number"`
	Metrics        []MetricDatapoints `json:"metrics" bson:"metrics"`
}

type HostMetrics struct {
	Host                   string                   `json:"host" bson:"host"`
	SystemMetrics          []MetricDatapoints       `json:"system_metrics" bson:"system_metrics"`
	ServiceInstanceMetrics []ServiceInstanceMetrics `json:"service_instance_metrics" bson:"service_instance_metrics"`
}

// HostMetricsMap is a map of host metrics
// The key is the host identifier
type HostMetricsMap map[string]MetricsMap
type MetricsMap struct {
	// the string depicts the full metrics identifier with metric(age){state}
	HostMetrics map[string]float64
	// the string is the full service identifier with map[job_name.instance.instance_number]map[metric(age){state}]float64
	ServiceInstanceMetrics map[string]map[string]float64
}

func (m *HostMetricsMap) InstanceMetricsForEvaluation(serviceIdentifier string) map[string]interface{} {
	values := make(map[string]interface{})

	// Aggregate metrics from all hosts
	for _, hostMetrics := range *m {
		// Add relevant host metrics - these are system-level metrics that evaluators might need
		for metricKey, metricValue := range hostMetrics.HostMetrics {
			// Only add if not already present (avoid duplicates from multiple hosts)
			if _, exists := values[metricKey]; !exists {
				values[metricKey] = metricValue
			}
		}

		// Add the service instance metrics if they exist
		if instanceMetrics, exists := hostMetrics.ServiceInstanceMetrics[serviceIdentifier]; exists {
			for metricKey, metricValue := range instanceMetrics {
				values[metricKey] = metricValue
			}
		}
	}

	return values
}

// MetricsBroker provides direct communication between processors
type MetricsBroker interface {
	// PublishMetrics publishes metrics for a job
	PublishMetrics(jobName string, metrics HostMetricsMap)
	// GetMetrics retrieves the latest metrics for a job as HostMetricsMap
	GetMetrics(jobName string) (HostMetricsMap, bool)
	// Subscribe to metrics updates for a job
	Subscribe(jobName string, callback func(HostMetricsMap))
	// Clear old metrics
	Cleanup(maxAge time.Duration)
}

// InMemoryMetricsBroker implements MetricsBroker using in-memory storage
type InMemoryMetricsBroker struct {
	mu          sync.RWMutex
	metrics     map[string]HostMetricsMap
	timestamps  map[string]time.Time
	subscribers map[string][]func(HostMetricsMap)
}

// NewInMemoryMetricsBroker creates a new in-memory metrics broker
func NewInMemoryMetricsBroker() MetricsBroker {
	return &InMemoryMetricsBroker{
		metrics:     make(map[string]HostMetricsMap),
		timestamps:  make(map[string]time.Time),
		subscribers: make(map[string][]func(HostMetricsMap)),
	}
}

func (b *InMemoryMetricsBroker) PublishMetrics(jobName string, metrics HostMetricsMap) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.metrics[jobName] = metrics
	b.timestamps[jobName] = time.Now()

	// Notify subscribers immediately for real-time processing
	if callbacks, exists := b.subscribers[jobName]; exists {
		for _, callback := range callbacks {
			go callback(metrics) // Run callbacks asynchronously to avoid blocking
		}
	}

	// Also notify wildcard subscribers (subscribed to all jobs)
	if callbacks, exists := b.subscribers["*"]; exists {
		for _, callback := range callbacks {
			go callback(metrics) // Run callbacks asynchronously
		}
	}
}

func (b *InMemoryMetricsBroker) GetMetrics(jobName string) (HostMetricsMap, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	metrics, exists := b.metrics[jobName]
	return metrics, exists
}

func (b *InMemoryMetricsBroker) Subscribe(jobName string, callback func(HostMetricsMap)) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.subscribers[jobName]; !exists {
		b.subscribers[jobName] = make([]func(HostMetricsMap), 0)
	}
	b.subscribers[jobName] = append(b.subscribers[jobName], callback)
}

func (b *InMemoryMetricsBroker) Cleanup(maxAge time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	for jobName, timestamp := range b.timestamps {
		if timestamp.Before(cutoff) {
			delete(b.metrics, jobName)
			delete(b.timestamps, jobName)
			delete(b.subscribers, jobName)
		}
	}
}

// Helper functions for metric ID building
func buildMetricID(name string, state string, age int) string {
	return name + "(" + fmt.Sprintf("%d", age) + ")" + "{" + state + "}"
}

func calculateAge(length int, index int) int {
	return (length - 1) - index
}

// Global metrics broker instance
var (
	globalBroker     MetricsBroker
	globalBrokerOnce sync.Once
)

// GetGlobalMetricsBroker returns the global metrics broker instance
func GetGlobalMetricsBroker() MetricsBroker {
	globalBrokerOnce.Do(func() {
		globalBroker = NewInMemoryMetricsBroker()

		// Start cleanup routine
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()

			for range ticker.C {
				globalBroker.Cleanup(10 * time.Minute) // Clean up metrics older than 10 minutes
			}
		}()
	})
	return globalBroker
}
