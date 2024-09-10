package backendexporter // import github.com/smnzlnsk/opentelemetry-components/exporters/backend

import (
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type marshaler struct {
	metricsMarshaler pmetric.Marshaler
}

func newMarshaler() (*marshaler, error) {
	var metricsMarshaler pmetric.Marshaler = &pmetric.ProtoMarshaler{}

	m := marshaler{
		metricsMarshaler: metricsMarshaler,
	}
	return &m, nil
}
