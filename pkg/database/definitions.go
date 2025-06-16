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

	if len(values) == 0 {
		availableServices := make([]string, 0)
		for _, hostMetrics := range *m {
			for serviceID := range hostMetrics.ServiceInstanceMetrics {
				availableServices = append(availableServices, serviceID)
			}
		}
	}

	return values
}

// MetricsBroker provides direct communication between processors
type MetricsBroker interface {
	// PublishMetrics publishes metrics for a job
	PublishMetrics(jobName string, metrics HostMetrics)
	// GetMetrics retrieves the latest metrics for a job
	GetMetrics(jobName string) (HostMetrics, bool)
	// GetMetricsAsMap retrieves the latest metrics for a job as a map
	GetMetricsAsMap(jobName string) (HostMetricsMap, bool)
	// Subscribe to metrics updates for a job
	Subscribe(jobName string, callback func(HostMetrics))
	// Clear old metrics
	Cleanup(maxAge time.Duration)
}

// InMemoryMetricsBroker implements MetricsBroker using in-memory storage
type InMemoryMetricsBroker struct {
	mu          sync.RWMutex
	metrics     map[string]HostMetrics
	timestamps  map[string]time.Time
	subscribers map[string][]func(HostMetrics)
}

// NewInMemoryMetricsBroker creates a new in-memory metrics broker
func NewInMemoryMetricsBroker() MetricsBroker {
	return &InMemoryMetricsBroker{
		metrics:     make(map[string]HostMetrics),
		timestamps:  make(map[string]time.Time),
		subscribers: make(map[string][]func(HostMetrics)),
	}
}

func (b *InMemoryMetricsBroker) PublishMetrics(jobName string, metrics HostMetrics) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.metrics[jobName] = metrics
	b.timestamps[jobName] = time.Now()

	// Notify subscribers
	if callbacks, exists := b.subscribers[jobName]; exists {
		for _, callback := range callbacks {
			go callback(metrics) // Run callbacks asynchronously
		}
	}
}

func (b *InMemoryMetricsBroker) GetMetrics(jobName string) (HostMetrics, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	metrics, exists := b.metrics[jobName]
	return metrics, exists
}

func (b *InMemoryMetricsBroker) GetMetricsAsMap(jobName string) (HostMetricsMap, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	metrics, exists := b.metrics[jobName]
	if !exists {
		return HostMetricsMap{}, false
	}

	// Transform to map format
	hostMetrics := make(HostMetricsMap)
	hostMetrics[metrics.Host] = MetricsMap{
		HostMetrics:            make(map[string]float64),
		ServiceInstanceMetrics: make(map[string]map[string]float64),
	}

	// Process system metrics
	for _, systemMetric := range metrics.SystemMetrics {
		for i, datapoint := range systemMetric.Datapoints {
			age := calculateAge(len(systemMetric.Datapoints), i)
			metricID := buildMetricID(systemMetric.Identifier.Name, systemMetric.Identifier.State, age)
			hostMetrics[metrics.Host].HostMetrics[metricID] = datapoint.Value
		}
	}

	// Process service instance metrics
	for _, serviceInstance := range metrics.ServiceInstanceMetrics {
		serviceID := serviceInstance.JobName + ".instance." + fmt.Sprintf("%d", serviceInstance.InstanceNumber)

		if _, exists := hostMetrics[metrics.Host].ServiceInstanceMetrics[serviceID]; !exists {
			hostMetrics[metrics.Host].ServiceInstanceMetrics[serviceID] = make(map[string]float64)
		}

		for _, metric := range serviceInstance.Metrics {
			for i, datapoint := range metric.Datapoints {
				age := calculateAge(len(metric.Datapoints), i)
				metricID := buildMetricID(metric.Identifier.Name, metric.Identifier.State, age)
				hostMetrics[metrics.Host].ServiceInstanceMetrics[serviceID][metricID] = datapoint.Value
			}
		}
	}

	return hostMetrics, true
}

func (b *InMemoryMetricsBroker) Subscribe(jobName string, callback func(HostMetrics)) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.subscribers[jobName]; !exists {
		b.subscribers[jobName] = make([]func(HostMetrics), 0)
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
