package cc

import (
	"context"
	"fmt"
	"os"
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
	Context           context.Context
	UI                *clui.UI
	NoCheck           bool
	DefaultConfigFile string
}

	UI      *clui.UI
	NoCheck bool
}

type runStatus int

const (
	StatusNotRunning runStatus = iota
	StatusRunning
)

func NewCommonContext(noCheck bool, defaultConfig string) (*CommonCtx, error) {
	return &CommonCtx{
		Context:           context.Background(),
		UI:                iostream.NewUI(iostream.DefaultIO()),
		NoCheck:           noCheck,
		DefaultConfigFile: defaultConfig,
	}, nil
}

func (c *CommonCtx) CheckRunStatus(containerName string, expectedStatus runStatus) bool {
	if c.NoCheck {
		return false
	}

	// set default container name if not specified.
	if containerName == "" {
		containerName = ContainerName()
	}

	running, err := dockerx.IsRunning(containerName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	}

	return lo.Ternary(running, StatusRunning, StatusNotRunning) == expectedStatus
}

func (c *CommonCtx) IsServing(grpcAddress string) bool {
	if c.NoCheck {
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
