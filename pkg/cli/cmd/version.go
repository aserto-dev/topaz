package cmd

import (
	"fmt"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/aserto-dev/topaz/pkg/cli/x"
	"github.com/aserto-dev/topaz/pkg/version"
)

type VersionCmd struct {
	Container         bool   `short:"c" help:"display topazd container version" default:"false"`
	ContainerRegistry string `optional:"" default:"${container_registry}" env:"CONTAINER_REGISTRY" help:"container registry (host[:port]/repo)"`
	ContainerImage    string `optional:"" default:"${container_image}" env:"CONTAINER_IMAGE" help:"container image name"`
	ContainerTag      string `optional:"" default:"${container_tag}" env:"CONTAINER_TAG" help:"container tag"`
	ContainerPlatform string `optional:"" default:"${container_platform}" env:"CONTAINER_PLATFORM" help:"container platform"`
	ContainerVersion  string `optional:"" hidden:"" default:"" env:"CONTAINER_VERSION"`
}

func (cmd *VersionCmd) Run(c *cc.CommonCtx) error {
	cmd.ContainerTag = cc.ContainerVersionTag(cmd.ContainerVersion, cmd.ContainerTag)

	fmt.Fprintf(c.StdOut(), "%s %s\n",
		x.AppName,
		version.GetInfo().String(),
	)

	if !cmd.Container {
		return nil
	}

	dc, err := dockerx.New()
	if err != nil {
		return err
	}

	image := cc.Container(
		cmd.ContainerRegistry, // registry
		cmd.ContainerImage,    // image
		cmd.ContainerTag,      // tag
	)

	if !dc.ImageExists(image) {
		fmt.Fprintf(c.StdOut(), "!!! container image %q does not exist locally\n", image)
		fmt.Fprint(c.StdOut(), "!!! run 'topaz install' to download\n")
		return nil
	}

	if err := dc.Run(
		dockerx.WithContainerImage(image),
		dockerx.WithEntrypoint([]string{"/app/topazd", "version"}),
		dockerx.WithContainerPlatform(cmd.ContainerPlatform),
		dockerx.WithContainerName("topazd-version"),
		dockerx.WithOutput(c.StdOut()),
		dockerx.WithError(c.StdErr()),
	); err != nil {
		return err
	}

	fmt.Fprintf(c.StdOut(), "\n")

	return nil
}
