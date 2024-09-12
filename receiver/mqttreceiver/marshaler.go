package mqttreceiver // import github.com/smnzlnsk/opentelemetry-components/receiver/mqttreceiver

import (
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type marshaler struct {
	metricsMarshaler   pmetric.Marshaler
	metricsUnmarshaler pmetric.Unmarshaler
}

func newMarshaler(encoding string) (*marshaler, error) {
	var metricsMarshaler pmetric.Marshaler
	var metricsUnmarshaler pmetric.Unmarshaler

	switch encoding {
	case "json":
		metricsMarshaler = &pmetric.JSONMarshaler{}
		metricsUnmarshaler = &pmetric.JSONUnmarshaler{}
	case "proto":
		metricsMarshaler = &pmetric.ProtoMarshaler{}
		metricsUnmarshaler = &pmetric.ProtoUnmarshaler{}
	}

	m := marshaler{
		metricsMarshaler:   metricsMarshaler,
		metricsUnmarshaler: metricsUnmarshaler,
	}
	return &m, nil
}
