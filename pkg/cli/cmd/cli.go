package cmd

import (
	"errors"
)

var (
	ErrNotRunning = errors.New("topaz is not running, use 'topaz start' or 'topaz run' to start")
	ErrIsRunning  = errors.New("topaz is already running, use 'topaz stop' to stop")
	ErrNotServing = errors.New("topaz gRPC endpoint not SERVING")
)

type FormatVersion int

const (
	V2 FormatVersion = 2
	V3 FormatVersion = 3
)

type CLI struct {
	Start             StartCmd     `cmd:"" help:"start topaz in daemon mode"`
	Stop              StopCmd      `cmd:"" help:"stop topaz instance"`
	Status            StatusCmd    `cmd:"" help:"status of topaz daemon process"`
	Run               RunCmd       `cmd:"" help:"run topaz in console mode"`
	Manifest          ManifestCmd  `cmd:"" help:"manifest commands"`
	Test              TestCmd      `cmd:"" help:"test assertions commands"`
	Templates         TemplateCmd  `cmd:"" help:"template commands"`
	Console           ConsoleCmd   `cmd:"" help:"open console in the browser"`
	Import            ImportCmd    `cmd:"" help:"import directory objects"`
	Export            ExportCmd    `cmd:"" help:"export directory objects"`
	Backup            BackupCmd    `cmd:"" help:"backup directory data"`
	Restore           RestoreCmd   `cmd:"" help:"restore directory data"`
	Install           InstallCmd   `cmd:"" help:"install topaz container"`
	Configure         ConfigCmd    `cmd:"" help:"configure topaz service"`
	Certs             CertsCmd     `cmd:"" help:"cert commands"`
	Update            UpdateCmd    `cmd:"" help:"update topaz container version"`
	Uninstall         UninstallCmd `cmd:"" help:"uninstall topaz container"`
	Load              LoadCmd      `cmd:"" help:"load manifest from file (DEPRECATED)"`
	Save              SaveCmd      `cmd:"" help:"save manifest to file  (DEPRECATED)"`
	Version           VersionCmd   `cmd:"" help:"version information"`
	NoCheck           bool         `name:"no-check" json:"noCheck,omitempty" short:"N" env:"TOPAZ_NO_CHECK" help:"disable local container status check"`
	DefaultConfigFile string       `name:"config" json:"configFile,omitempty" default:"config.yaml" short:"c" help:"use topaz configuration file"`
}
