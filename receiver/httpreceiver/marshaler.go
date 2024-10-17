package httpreceiver // import github.com/smnzlnsk/opentelemetry-components/receiver/httpreceiver

import (
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type marshaler struct {
	metricsMarshaler   pmetric.Marshaler
	metricsUnmarshaler pmetric.Unmarshaler
}

func newMarshaler() (*marshaler, error) {
	return &marshaler{
		metricsMarshaler:   &pmetric.ProtoMarshaler{},
		metricsUnmarshaler: &pmetric.ProtoUnmarshaler{},
	}, nil
}
