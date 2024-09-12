package mqttexporter // import github.com/smnzlnsk/opentelemetry-components/exporters/mqttexporter

import (
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type marshaler struct {
	metricsMarshaler pmetric.Marshaler
}

func newMarshaler(encoding string) (*marshaler, error) {
	var metricsMarshaler pmetric.Marshaler
	switch encoding {
	case "json":
		metricsMarshaler = &pmetric.JSONMarshaler{}
	case "proto":
		metricsMarshaler = &pmetric.ProtoMarshaler{}
	}

	m := marshaler{
		metricsMarshaler: metricsMarshaler,
	}
	return &m, nil
}
