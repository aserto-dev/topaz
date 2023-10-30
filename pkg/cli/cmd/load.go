package cmd

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/pkg/errors"
)

type LoadCmd struct {
	Path string `arg:"" type:"existingfile" default:"manifest.yaml" help:"absolute path to manifest file"`
	clients.Config
}

func (cmd *LoadCmd) Run(c *cc.CommonCtx) error {
	return errors.Errorf("The \"topaz load\" command has been deprecated, see \"topaz manifest set\".")
}
