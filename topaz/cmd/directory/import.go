package directory

import dsc "github.com/aserto-dev/topaz/topaz/clients/directory"

type ImportCmd struct {
	dsc.Config

	File string `flag:"file" short:"f" type:"path" help:"path to source import file"`
}
