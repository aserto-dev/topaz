package directory

import dsc "github.com/aserto-dev/topaz/topaz/clients/directory"

type RestoreCmd struct {
	dsc.Config

	File string `flag:"file" short:"f" type:"path" help:"path to target file (jsonl)"`
}
