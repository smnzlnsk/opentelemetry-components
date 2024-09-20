package mqttreceiver // import github.com/smnzlnsk/opentelemetry-components/receiver/mqttreceiver

import (
	"errors"
	"os"
	"strconv"
	"time"
)

// Config represents the receiver config settings within the collector's config.yaml
type Config struct {
	Interval string       `mapstructure:"interval"`
	ClientID string       `mapstructure:"client_id"`
	Topic    string       `mapstructure:"topic"`
	Encoding string       `mapstructure:"encoding"`
	Broker   BrokerConfig `mapstructure:"broker"`
}

type BrokerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// Validate checks if the receiver configuration is valid
// Additionally it overrides the configuration set if the respective environment variables are set
func (cfg *Config) Validate() error {
	// client ID
	if os.Getenv("MQTTRECEIVER_MONITORING_CLIENT_ID") != "" {
		cfg.ClientID = os.Getenv("MQTTRECEIVER_CLIENT_ID")
	}
	if cfg.ClientID == "" {
		return errors.New("client id has to be set")
	}

	// topic
	if cfg.Topic == "" {
		return errors.New("topic has to be set")
	}

	// interval
	interval, _ := time.ParseDuration(cfg.Interval)
	if interval.Seconds() < 1 {
		return errors.New("interval cannot be sub-second")
	}
	if interval.Seconds() > 60 {
		return errors.New("interval cannot be more than a minute")
	}

	// encoding
	if cfg.Encoding != "json" && cfg.Encoding != "proto" {
		return errors.New("unsupported encoding. supported encoding: json, proto")
	}

	// broker.host
	if os.Getenv("MQTTRECEIVER_MONITORING_BROKER_HOST") != "" {
		cfg.Broker.Host = os.Getenv("MQTTRECEIVER_MONITORING_BROKER_HOST")
	}
	//if !isValidIP(cfg.Broker.Host) {
	//	return errors.New("broker.host is invalid")
	//}

	// broker.port
	if os.Getenv("MQTTRECEIVER_MONITORING_BROKER_PORT") != "" {
		if v, err := strconv.Atoi(os.Getenv("MQTTRECEIVER_MONITORING_BROKER_PORT")); err != nil {
			cfg.Broker.Port = v
		}
	}
	if cfg.Broker.Port > 65535 || cfg.Broker.Port < 1 {
		return errors.New("broker.port is invalid")
	}
	return nil
}
