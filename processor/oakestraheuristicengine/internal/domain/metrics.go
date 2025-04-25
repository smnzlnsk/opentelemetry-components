package domain

import "time"

type MetricKey struct {
	Name  string          `json:"name" bson:"name"`
	State string          `json:"state" bson:"state"`
	Type  MetricValueType `json:"type" bson:"type"`
}

type MetricDatapoint struct {
	Value     float64   `json:"value" bson:"value"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

type DBMetricDatapoint struct {
	Identifier MetricKey         `json:"identifier" bson:"identifier"`
	Datapoints []MetricDatapoint `json:"datapoints" bson:"datapoints"`
}

type DBServiceInstanceMetrics struct {
	JobName        string              `json:"job_name" bson:"job_name"`
	InstanceNumber int                 `json:"instance_number" bson:"instance_number"`
	Metrics        []DBMetricDatapoint `json:"metrics" bson:"metrics"`
}

type DBHostMetrics struct {
	Host                   string                     `json:"host" bson:"host"`
	SystemMetrics          []DBMetricDatapoint        `json:"system_metrics" bson:"system_metrics"`
	ServiceInstanceMetrics []DBServiceInstanceMetrics `json:"service_instance_metrics" bson:"service_instance_metrics"`
}

type MapHostMetrics map[string]MetricsStruct
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

type ServiceInstanceMetricsMap map[string]float64 // the string depicts the full metrics identifier with <metric_name>|<metric_state>|<datapoint_number>

type MetricValueType int

const (
	MetricValueTypeRaw MetricValueType = iota
	MetricValueTypeSum
	MetricValueTypeAvg
	MetricValueTypeCount
	MetricValueTypeMin
	MetricValueTypeMax
	MetricValueTypeStdDev
	MetricValueTypeMedian
	MetricValueTypePercentile90
	MetricValueTypePercentile95
	MetricValueTypePercentile99
	MetricValueTypePercentile999
	MetricValueTypePercentile9999
)

func (t MetricValueType) String() string {
	return [...]string{
		"raw",
		"sum",
		"avg",
		"count",
		"min",
		"max",
		"stddev",
		"median",
		"p90",
		"p95",
		"p99",
		"p999",
		"p9999",
	}[t]
}
