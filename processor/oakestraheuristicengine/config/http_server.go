package config

// HTTPServerConfig defines the configuration for the HTTP server
type HTTPServerConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Host    string `mapstructure:"host"`
}
