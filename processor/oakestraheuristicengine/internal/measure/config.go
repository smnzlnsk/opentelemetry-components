package measure

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

type MeasureConfig interface {
	Validate() error
}

type measuresConfig struct {
	measures map[string]MeasureConfig `mapstructure:"-"`
}

var _ component.Config = (*measuresConfig)(nil)
var _ confmap.Unmarshaler = (*measuresConfig)(nil)

func (c *measuresConfig) Validate() error {
	return nil
}

func (c *measuresConfig) Unmarshal(cp *confmap.Conf) error {
	if cp == nil {
		return nil
	}

	err := cp.Unmarshal(c, confmap.WithIgnoreUnused())
	if err != nil {
		return err
	}

	c.measures = make(map[string]MeasureConfig)

	for key, value := range cp.ToStringMap() {
		c.measures[key] = value.(MeasureConfig)
	}

	return nil
}
