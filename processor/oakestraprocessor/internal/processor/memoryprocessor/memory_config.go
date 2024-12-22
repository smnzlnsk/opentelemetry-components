package memoryprocessor

import (
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/processor/memoryprocessor/internal/metadata"
)

type Config struct {
	metadata.MetricsBuilderConfig `mapstructure:",squash"`
	internal.Config
}
