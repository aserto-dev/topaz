package cc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/aserto-dev/topaz/internal/fs"
	"github.com/aserto-dev/topaz/topaz/dockerx"
	"github.com/docker/docker/api/types/container"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

var (
	ErrNotRunning = errors.New("topaz is not running, use 'topaz start' or 'topaz run' to start")
	ErrIsRunning  = errors.New("topaz is already running, use 'topaz stop' to stop")
)

type Config struct {
	Version  int            `json:"version"`
	Active   ActiveConfig   `json:"active"`
	Running  RunningConfig  `json:"running"`
	Defaults DefaultsConfig `json:"defaults"`
}

type ActiveConfig struct {
	Config     string `json:"config"`
	ConfigFile string `json:"config_file"`
}
type RunningConfig struct {
	ActiveConfig

	ContainerName string `json:"container_name"`
}

type DefaultsConfig struct {
	NoCheck           bool   `json:"no_check"`
	NoColor           bool   `json:"no_color"`
	ContainerRegistry string `json:"container_registry"`
	ContainerImage    string `json:"container_image"`
	ContainerTag      string `json:"container_tag"`
	ContainerPlatform string `json:"container_platform"`
}

type runStatus int

const (
	StatusNotRunning runStatus = iota
	StatusRunning
)

var (
	config   *Config
	defaults DefaultsConfig
	once     sync.Once
)

func NewConfig(ctx context.Context, noCheck bool, configFilePath string) (*Config, error) {
	cfg := &Config{
		Version: 1,
		Active: ActiveConfig{
			ConfigFile: filepath.Join(GetTopazCfgDir(), "config.yaml"),
			Config:     "config.yaml",
		},
		Defaults: DefaultsConfig{
			NoCheck: noCheck,
		},
		Running: RunningConfig{},
	}

	if _, err := os.Stat(configFilePath); err == nil {
		data, err := os.ReadFile(configFilePath)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	setDefaults(cfg)

	return cfg, nil
}

func GetConfig() *Config {
	if config == nil {
		panic("config not initialized")
	}

	return config
}

func (cfg *Config) CheckRunStatus(containerName string, expectedStatus runStatus) bool {
	if cfg.Defaults.NoCheck {
		return false
	}

	// set default container name if not specified.
	if containerName == "" {
		containerName = ContainerName(cfg.Active.ConfigFile)
	}

	dc, err := dockerx.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	}

	running, err := dc.IsRunning(containerName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	}

	return lo.Ternary(running, StatusRunning, StatusNotRunning) == expectedStatus
}

func (cfg *Config) GetRunningContainers() ([]container.Summary, error) {
	dc, err := dockerx.New()
	if err != nil {
		return nil, err
	}

	topazContainers, err := dc.GetRunningTopazContainers()
	if err != nil {
		return nil, err
	}

	return topazContainers, nil
}

func (cfg *Config) SaveContextConfig(configurationFile string) error {
	cliConfig := filepath.Join(GetTopazDir(), configurationFile)

	kongConfigBytes, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := os.WriteFile(cliConfig, kongConfigBytes, fs.FileModeOwnerRW); err != nil {
		return err
	}

	defaults = cfg.Defaults

	return nil
}

func setDefaults(cfg *Config) {
	once.Do(func() {
		config = cfg
		defaults = cfg.Defaults
	})
}
