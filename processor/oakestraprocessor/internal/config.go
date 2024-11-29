package internal

// Config is the configuration of a processor
type Config interface {
}

type ProcessorConfig struct {
	Formula map[int]Calculation `mapstructure:"formula"`
}
