package file

import (
	"os"
	"path/filepath"
)

const default_decision_log_filename string = "decisions.json"

type Config struct {
	LogFilePath   string `json:"log_file_path"`
	MaxFileSizeMB int    `json:"max_file_size_mb"`
	MaxFileCount  int    `json:"max_file_count"`
	Compress      bool   `json:"compress"`
}

func (cfg *Config) SetDefaults() {
	if cfg.LogFilePath == "" {
		pwd, err := os.Getwd()
		if err != nil {
			cfg.LogFilePath = filepath.Join(".", default_decision_log_filename)
		} else {
			cfg.LogFilePath = filepath.Join(pwd, default_decision_log_filename)
		}
	}

	if cfg.MaxFileSizeMB == 0 {
		cfg.MaxFileSizeMB = 50
	}

	if cfg.MaxFileCount == 0 {
		cfg.MaxFileCount = 2
	}
}
