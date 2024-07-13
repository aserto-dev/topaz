package cc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aserto-dev/topaz/pkg/cli/cc/iostream"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/docker/docker/api/types"
	"github.com/fatih/color"
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

// ConMsg - console message, send either StdErr or StdOut.
type ConMsg struct {
	out   io.Writer
	color *color.Color
}

// Con() - console message send to StdErr, default no-color.
func (c *CommonCtx) Con() *ConMsg {
	return &ConMsg{
		out:   color.Error,
		color: color.New(),
	}
}

// Out() - console output message send to StdOut, default no-color.
func (c *CommonCtx) Out() *ConMsg {
	return &ConMsg{
		out:   color.Output,
		color: color.New(),
	}
}

// Info() - info console message (green).
func (cm *ConMsg) Info() *ConMsg {
	cm.color.Add(color.FgGreen)
	return cm
}

// Warn() - warning console message (yellow).
func (cm *ConMsg) Warn() *ConMsg {
	cm.color.Add(color.FgYellow)
	return cm
}

// Error() - error console message (red).
func (cm *ConMsg) Error() *ConMsg {
	cm.color.Add(color.FgRed)
	return cm
}

// Msg() - sends the con|out message, by default adds a CrLr when not present.
func (cm *ConMsg) Msg(message string, args ...interface{}) {
	color.NoColor = NoColor()

	if !strings.HasSuffix(message, "\n") {
		message += "\n"
	}

	if len(args) == 0 {
		cm.color.Fprint(color.Error, message)
		return
	}

	cm.color.Fprintf(color.Error, message, args...)
}
