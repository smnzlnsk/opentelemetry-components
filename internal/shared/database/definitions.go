package database

import (
	"time"

	"github.com/smnzlnsk/opentelemetry-components/internal/shared/metric"
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

type MapHostMetrics map[string]MetricsStruct
type ServiceInstanceMetricsMap map[string]float64 // the string depicts the full metrics identifier with <metric_name>|<metric_state>|<datapoint_number>
type MetricsStruct struct {
	// the string depicts the full metrics identifier with <metric_name>|<metric_state>|<datapoint_number>
	HostMetrics map[string]float64
	// the string is the full service identifier with <job_name>.instance.<instance_number>
	ServiceInstanceMetrics map[string]ServiceInstanceMetricsMap
}

func (m *MapHostMetrics) InstanceMetricsForEvaluation(serviceIdentifier string) map[string]interface{} {
	values := make(map[string]interface{})
	for _, hostMetrics := range *m {
		// Only include host metrics if there's a match for the serviceIdentifier
		if instanceMetrics, exists := hostMetrics.ServiceInstanceMetrics[serviceIdentifier]; exists {
			// Add host metrics when there's a service match
			for metricKey, metricValue := range hostMetrics.HostMetrics {
				values[metricKey] = metricValue
			}

			// Add the service instance metrics
			for metricKey, metricValue := range instanceMetrics {
				values[metricKey] = metricValue
			}
			break
		}
	}
	return values
}
