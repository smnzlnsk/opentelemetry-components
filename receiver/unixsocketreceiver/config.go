package unixsocketreceiver // import github.com/smnzlnsk/opentelemetry-components/receiver/unixsocketreceiver

import (
	"errors"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Folder   string `mapstructure:"folder"`
	Interval string `mapstructure:"polling_interval"`
}

func (cfg *Config) Validate() error {
	// interval
	interval, _ := time.ParseDuration(cfg.Interval)
	if interval.Seconds() < 1 {
		return errors.New("interval cannot be sub-second")
	}
	if interval.Seconds() > 60 {
		return errors.New("interval cannot be more than a minute")
	}

	// folder
	if cfg.Folder == "" {
		return errors.New("folder is required")
	}
	if !filepath.IsAbs(cfg.Folder) {
		return errors.New("folder path must be absolute")
	}

	info, err := os.Stat(cfg.Folder)
	// create folder if it does not exist
	if errors.Is(err, os.ErrNotExist) {
		dirErr := os.Mkdir(cfg.Folder, 0755)
		if dirErr != nil {
			return dirErr
		}
	} else if err != nil {
		return err
	}

	if !info.IsDir() {
		return errors.New("folder is not a directory")
	}

	return nil
}
