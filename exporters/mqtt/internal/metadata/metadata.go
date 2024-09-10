package metadata

import (
	"go.opentelemetry.io/collector/component"
)

var (
	Type      = component.MustNewType("mqttexporter")
	ScopeName = "github.com/smnzlnsk/opentelemetry-components/exporters/mqtt"
)
