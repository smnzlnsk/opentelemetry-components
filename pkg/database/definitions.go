package database

import (
	"fmt"
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

func (m *HostMetricsMap) String() string {
	for _, metrics := range *m {
		fmt.Println(metrics.String())
	}
	return ""
}

func (m *MetricsMap) String() string {
	for metricKey, metricValue := range m.HostMetrics {
		fmt.Printf("\t%s: %f\n", metricKey, metricValue)
	}
	for serviceInstance, metrics := range m.ServiceInstanceMetrics {
		fmt.Printf("\t%s:\n", serviceInstance)
		for metricKey, metricValue := range metrics {
			fmt.Printf("\t\t%s: %f\n", metricKey, metricValue)
		}
	}
	return ""
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

// Helper functions for metric ID building
func BuildMetricID(name string, state string, age int) string {
	return name + "(" + fmt.Sprintf("%d", age) + ")" + "{" + state + "}"
}

func CalculateAge(length int, index int) int {
	return (length - 1) - index
}
