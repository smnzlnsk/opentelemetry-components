package metadata

import (
	"go.opentelemetry.io/collector/component"
)

var (
	Type      = component.MustNewType("backend")
	ScopeName = "github.com/smnzlnsk/opentelemetry-components/exporter/backend"
)
