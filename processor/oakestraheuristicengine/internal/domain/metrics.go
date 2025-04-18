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
