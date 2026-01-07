package file

import (
	"os"
)

type Config struct {
	LogFilePath   string `json:"log_file_path"`
	MaxFileSizeMB int    `json:"max_file_size_mb"`
	MaxFileCount  int    `json:"max_file_count"`
}

func (cfg *Config) SetDefaults() {
	if cfg.LogFilePath == "" {
		pwd, err := os.Getwd()
		if err != nil {
			cfg.LogFilePath = "."
		} else {
			cfg.LogFilePath = pwd
		}
	}

	if cfg.MaxFileSizeMB == 0 {
		cfg.MaxFileSizeMB = 50
	}

	if cfg.MaxFileCount == 0 {
		cfg.MaxFileCount = 2
	}
}
