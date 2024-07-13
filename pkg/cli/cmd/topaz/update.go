package topaz

import (
	"fmt"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/fatih/color"
)

type UpdateCmd struct {
	ContainerRegistry string `optional:"" default:"${container_registry}" env:"CONTAINER_REGISTRY" help:"container registry (host[:port]/repo)"`
	ContainerImage    string `optional:"" default:"${container_image}" env:"CONTAINER_IMAGE" help:"container image name"`
	ContainerTag      string `optional:"" default:"${container_tag}" env:"CONTAINER_TAG" help:"container tag"`
	ContainerPlatform string `optional:"" default:"${container_platform}" env:"CONTAINER_PLATFORM" help:"container platform"`
	ContainerVersion  string `optional:"" hidden:"" default:"" env:"CONTAINER_VERSION"`
}

func (cmd *UpdateCmd) Run(c *cc.CommonCtx) error {
	cmd.ContainerTag = cc.ContainerVersionTag(cmd.ContainerVersion, cmd.ContainerTag)

	fmt.Fprint(c.StdOut(), color.GreenString(">>> updating %s (%s)...\n",
		cc.Container(
			cmd.ContainerRegistry, // registry
			cmd.ContainerImage,    // image
			cmd.ContainerTag,      // tag
		),
		cmd.ContainerPlatform,
	))

	dc, err := dockerx.New()
	if err != nil {
		return err
	}

	return dc.PullImage(
		cc.Container(
			cmd.ContainerRegistry, // registry
			cmd.ContainerImage,    // image
			cmd.ContainerTag,      // tag
		),
		cmd.ContainerPlatform,
	)
}
