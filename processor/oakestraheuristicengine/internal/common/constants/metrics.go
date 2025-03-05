package constants

import "github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"

const (
	MetricValueTypeRaw types.MetricValueType = iota
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
