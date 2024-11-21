package internal

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type MetricDatapoint struct {
	Metadata MetricMetadata
	Value    Datapoint
}

func CreateMetricDatapoint(metric pmetric.Metric, idx int) *MetricDatapoint {
	ndp := metric.Sum().DataPoints().At(idx)
	md := &MetricDatapoint{
		Metadata: MetricMetadata{
			MetricType: metric.Type(),
			MetricName: metric.Name(),
			MetricUnit: metric.Unit(),
			Attributes: metric.Metadata(),
		},
		Value: Datapoint{
			ValueDataType: ndp.ValueType(),
			FloatValue:    ndp.DoubleValue(),
		},
	}
	return md
}

type MetricMetadata struct {
	MetricType pmetric.MetricType
	MetricName string
	MetricUnit string
	Attributes pcommon.Map
}

type Datapoint struct {
	ValueDataType pmetric.NumberDataPointValueType
	FloatValue    float64
}
