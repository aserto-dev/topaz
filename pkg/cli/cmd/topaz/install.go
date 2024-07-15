package topaz

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
)

type InstallCmd struct {
	ContainerRegistry string `optional:"" default:"${container_registry}" env:"CONTAINER_REGISTRY" help:"container registry (host[:port]/repo)"`
	ContainerImage    string `optional:"" default:"${container_image}" env:"CONTAINER_IMAGE" help:"container image name"`
	ContainerTag      string `optional:"" default:"${container_tag}" env:"CONTAINER_TAG" help:"container tag"`
	ContainerPlatform string `optional:"" default:"${container_platform}" env:"CONTAINER_PLATFORM" help:"container platform"`
	ContainerVersion  string `optional:"" hidden:"" default:"" env:"CONTAINER_VERSION"`
}

func (cmd *InstallCmd) Run(c *cc.CommonCtx) error {
	cmd.ContainerTag = cc.ContainerVersionTag(cmd.ContainerVersion, cmd.ContainerTag)

	if c.CheckRunStatus(cc.ContainerName(c.Config.Active.Config), cc.StatusRunning) {
		return cc.ErrIsRunning
	}

	c.Con().Info().Msg(">>> installing %s (%s)...",
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
