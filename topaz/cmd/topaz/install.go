package topaz

import (
	"context"

	"github.com/aserto-dev/topaz/topaz/cc"
	"github.com/aserto-dev/topaz/topaz/dockerx"
)

type InstallCmd struct {
	ContainerRegistry string `optional:"" default:"${container_registry}" env:"CONTAINER_REGISTRY" help:"container registry (host[:port]/repo)"`
	ContainerImage    string `optional:"" default:"${container_image}" env:"CONTAINER_IMAGE" help:"container image name"`
	ContainerTag      string `optional:"" default:"${container_tag}" env:"CONTAINER_TAG" help:"container tag"`
	ContainerPlatform string `optional:"" default:"${container_platform}" env:"CONTAINER_PLATFORM" help:"container platform"`
	ContainerVersion  string `optional:"" hidden:"" default:"" env:"CONTAINER_VERSION"`
}

func (cmd *InstallCmd) Run(ctx context.Context) error {
	cmd.ContainerTag = cc.ContainerVersionTag(cmd.ContainerVersion, cmd.ContainerTag)

	cfg := cc.GetConfig()

	if cfg.CheckRunStatus(cc.ContainerName(cfg.Active.Config), cc.StatusRunning) {
		return cc.ErrIsRunning
	}

	cc.Con().Info().Msg(">>> installing %s (%s)...",
		cc.Container(
			cmd.ContainerRegistry, // registry
			cmd.ContainerImage,    // image name
			cmd.ContainerTag,      // tag
		),
		cmd.ContainerPlatform, // os/arch
	)

	dc, err := dockerx.New()
	if err != nil {
		return err
	}

	return dc.PullImage(
		cc.Container(
			cmd.ContainerRegistry, // registry
			cmd.ContainerImage,    // image name
			cmd.ContainerTag,      // tag
		),
		cmd.ContainerPlatform, // os/arch
	)
}
