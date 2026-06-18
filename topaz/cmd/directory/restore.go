package directory

import dsc "github.com/aserto-dev/topaz/topaz/clients/directory"

type RestoreCmd struct {
	dsc.Config

	File string `arg:"" default:"backup.tar.gz" help:"path to source backup file"`
}
