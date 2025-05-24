package metric

type Key struct {
	Name  string    `json:"name" bson:"name"`
	State string    `json:"state" bson:"state"`
	Type  ValueType `json:"type" bson:"type"`
}

type ValueType int

const (
	MetricValueTypeRaw ValueType = iota
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

func (t ValueType) String() string {
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
