package cpuprocessor

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/processor/cpuprocessor/internal/metadata"
)

type Config struct {
	metadata.MetricsBuilderConfig `mapstructure:",squash"`
	internal.Config
}
