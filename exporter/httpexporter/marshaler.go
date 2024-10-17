package httpexporter // import github.com/smnzlnsk/opentelemetry-components/exporter/httpexporter

import (
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type marshaler struct {
	metricsMarshaler pmetric.Marshaler
}

func newMarshaler() (*marshaler, error) {
	return &marshaler{
		metricsMarshaler: &pmetric.ProtoMarshaler{},
	}, nil
}
