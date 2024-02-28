package cc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aserto-dev/clui"
	"github.com/aserto-dev/topaz/pkg/cli/cc/iostream"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/fullstorydev/grpcurl"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type CommonCtx struct {
	Context context.Context
	UI      *clui.UI
	Config  *CLIConfig
}

type CLIConfig struct {
	NoCheck           bool
	TopazConfigFile   string
	ContainerName     string
	ContainerRegistry string
	ContainerImage    string
	ContainerTag      string
}

type runStatus int

const (
	StatusNotRunning runStatus = iota
	StatusRunning
)

func NewCommonContext(noCheck bool, configFilePath string) (*CommonCtx, error) {
	ctx := &CommonCtx{
		Context: context.Background(),
		UI:      iostream.NewUI(iostream.DefaultIO()),
		Config: &CLIConfig{
			NoCheck:           noCheck,
			TopazConfigFile:   filepath.Join(GetTopazCfgDir(), "config.yaml"),
			ContainerName:     ContainerName(filepath.Join(GetTopazCfgDir(), "config.yaml")),
			ContainerRegistry: ContainerRegistry(),
			ContainerImage:    ContainerImage(),
			ContainerTag:      ContainerTag(),
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

	return ctx, nil
}

func (c *CommonCtx) CheckRunStatus(containerName string, expectedStatus runStatus) bool {
	if c.Config.NoCheck {
		return false
	}

	// set default container name if not specified.
	if containerName == "" {
		containerName = ContainerName(c.Config.TopazConfigFile)
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

func (c *CommonCtx) IsServing(grpcAddress string) bool {
	if c.Config.NoCheck {
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
	err = os.WriteFile(cliConfig, kongConfigBytes, 0666) // nolint
	if err != nil {
		return err
	}
	return nil
}
