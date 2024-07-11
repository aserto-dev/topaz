package cc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aserto-dev/topaz/pkg/cli/cc/iostream"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/docker/docker/api/types"
	"github.com/fullstorydev/grpcurl"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ErrNotRunning = errors.New("topaz is not running, use 'topaz start' or 'topaz run' to start")
	ErrIsRunning  = errors.New("topaz is already running, use 'topaz stop' to stop")
	ErrNotServing = errors.New("topaz gRPC endpoint not SERVING")
)

type CommonCtx struct {
	Context context.Context
	std     *iostream.StdIO
	Config  *CLIConfig
}

type CLIConfig struct {
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
	defaults DefaultsConfig
	once     sync.Once
)

func NewCommonContext(noCheck bool, configFilePath string) (*CommonCtx, error) {
	ctx := &CommonCtx{
		Context: context.Background(),
		std:     iostream.DefaultIO(),
		Config: &CLIConfig{
			Version: 1,
			Active: ActiveConfig{
				ConfigFile: filepath.Join(GetTopazCfgDir(), "config.yaml"),
				Config:     "config.yaml",
			},
			Defaults: DefaultsConfig{
				NoCheck: noCheck,
			},
			Running: RunningConfig{},
		},
	}

	if _, err := os.Stat(configFilePath); err == nil {
		data, err := os.ReadFile(configFilePath)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(data, ctx.Config)
		if err != nil {
			return nil, err
		}
	}

	setDefaults(ctx)

	return ctx, nil
}

func (c *CommonCtx) CheckRunStatus(containerName string, expectedStatus runStatus) bool {
	if c.Config.Defaults.NoCheck {
		return false
	}

	// set default container name if not specified.
	if containerName == "" {
		containerName = ContainerName(c.Config.Active.ConfigFile)
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

func (c *CommonCtx) GetRunningContainers() ([]*types.Container, error) {
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

func (c *CommonCtx) IsServing(grpcAddress string) bool {
	if c.Config.Defaults.NoCheck {
		return true
	}

	tlsConf, err := grpcurl.ClientTLSConfig(true, "", "", "")
	if err != nil {
		return false
	}

	creds := credentials.NewTLS(tlsConf)

	opts := []grpc.DialOption{
		grpc.WithUserAgent("topaz/dev-build (no version set)"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = grpcurl.BlockingDial(ctx, "tcp", grpcAddress, creds, opts...)

	return err == nil
}

func (c *CommonCtx) SaveContextConfig(configurationFile string) error {
	cliConfig := filepath.Join(GetTopazDir(), configurationFile)

	kongConfigBytes, err := json.Marshal(c.Config)
	if err != nil {
		return err
	}
	err = os.WriteFile(cliConfig, kongConfigBytes, 0o666) // nolint
	if err != nil {
		return err
	}

	defaults = c.Config.Defaults
	return nil
}

func setDefaults(ctx *CommonCtx) {
	once.Do(func() {
		defaults = ctx.Config.Defaults
	})
}

func (c *CommonCtx) StdOut() io.Writer {
	return c.std.StdOut()
}

func (c *CommonCtx) StdErr() io.Writer {
	return c.std.StdErr()
}
