package config

type InterfacesConfig struct {
	Alert    InterfaceConfig `mapstructure:"alert"`
	Route    InterfaceConfig `mapstructure:"route"`
	Schedule InterfaceConfig `mapstructure:"schedule"`
}

type InterfaceConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}
