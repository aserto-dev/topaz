package impl

import (
	"context"
	"runtime"

	info "github.com/aserto-dev/go-grpc/aserto/common/info/v1"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/version"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/rs/zerolog"
)

// InfoServer internal - returns basic system information
type InfoServer struct {
	logger            *zerolog.Logger
	directoryResolver resolvers.DirectoryResolver
}

// NewInfoServer creates a new SystemServer instance
func NewInfoServer(logger *zerolog.Logger, cfg *config.Config, directoryResolver resolvers.DirectoryResolver) (*InfoServer, error) {
	newLogger := logger.With().Str("component", "api.system-server").Logger()

	return &InfoServer{
		logger:            &newLogger,
		directoryResolver: directoryResolver,
	}, nil
}

func (s *InfoServer) Info(ctx context.Context, req *info.InfoRequest) (*info.InfoResponse, error) {
	var (
		si *info.SystemInfo
		vi *info.VersionInfo
	)

	buildVersion := version.GetInfo()

	return &info.InfoResponse{
		System:  si,
		Version: vi,
		Build: &info.BuildInfo{
			Version: buildVersion.Version,
			Commit:  buildVersion.Commit,
			Date:    buildVersion.Date,
			Os:      runtime.GOOS,
			Arch:    runtime.GOARCH,
		},
	}, nil
}
