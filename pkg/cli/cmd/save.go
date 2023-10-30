package cmd

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/pkg/errors"
)

type SaveCmd struct {
	File string `arg:"" type:"path" default:"manifest.yaml" help:"absolute path to manifest file"`
	clients.Config
}

func (cmd *SaveCmd) Run(c *cc.CommonCtx) error {
	return errors.Errorf("The \"topaz save\" command has been deprecated, see \"topaz manifest get\".")
}
