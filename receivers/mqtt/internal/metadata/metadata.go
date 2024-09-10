package metadata

import (
	"go.opentelemetry.io/collector/component"
)

var (
	Type      = component.MustNewType("mqttreceiver")
	ScopeName = "github.com/smnzlnsk/opentelemetry-components/receivers/mqtt"
)
