package cmd

import dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"

type CLI struct {
	Init InitCmd `cmd:"" help:"create new database file"`
	Set  SetCmd  `cmd:"" help:"set manifest"`
	Load LoadCmd `cmd:"" help:"load data"`
	Sync SyncCmd `cmd:"" help:"sync data"`
}

type InitCmd struct {
	DBFile string `arg:"" help:"db file name"`
}

type SetCmd struct {
	DBFile   string `arg:"" help:"db file name" type:"existingfile"`
	Manifest string `arg:"" help:"manifest file path" type:"existingfile"`
}

type LoadCmd struct {
	DBFile  string `arg:"" help:"db file name" type:"existingfile"`
	DataDir string `arg:"" help:"data file directory" type:"existingdir"`
}

type SyncCmd struct {
	DBFile string   `arg:"" help:"db file name" type:"existingfile"`
	Mode   []string `flag:"" short:"m" enum:"manifest,full,diff,watermark" required:"" help:"sync mode"`
	dsc.Config
}
