package config

type ServiceManagerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}
