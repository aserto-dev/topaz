package cmd

import (
	"github.com/aserto-dev/topaz/topaz/cmd/access"
	"github.com/aserto-dev/topaz/topaz/cmd/authorizer"
	"github.com/aserto-dev/topaz/topaz/cmd/certs"
	"github.com/aserto-dev/topaz/topaz/cmd/configure"
	"github.com/aserto-dev/topaz/topaz/cmd/directory"
	"github.com/aserto-dev/topaz/topaz/cmd/templates"
	"github.com/aserto-dev/topaz/topaz/cmd/topaz"
)

type SaveContext bool

var Save SaveContext

//nolint:lll
type CLI struct {
	Start      topaz.StartCmd           `cmd:"" help:"start topaz instance (daemon mode)"`
	Stop       topaz.StopCmd            `cmd:"" help:"stop topaz instance"`
	Restart    topaz.RestartCmd         `cmd:"" help:"restart topaz instance"`
	Status     topaz.StatusCmd          `cmd:"" help:"status of topaz daemon process"`
	Config     configure.ConfigCmd      `cmd:"" help:"configure topaz instance"`
	Run        topaz.RunCmd             `cmd:"" help:"start topaz instance (console mode)"`
	Templates  templates.TemplateCmd    `cmd:"" help:"template commands"`
	Console    topaz.ConsoleCmd         `cmd:"" help:"open topaz console in the browser"`
	Directory  directory.DirectoryCmd   `cmd:"" aliases:"ds" help:"directory service commands"`
	Authorizer authorizer.AuthorizerCmd `cmd:"" aliases:"az" help:"authorizer service commands"`
	Access     access.AccessCmd         `cmd:"" aliases:"ac" help:"access service commands "`
	Certs      certs.CertsCmd           `cmd:"" help:"certificate management"`
	Install    topaz.InstallCmd         `cmd:"" help:"install topaz container"`
	Uninstall  topaz.UninstallCmd       `cmd:"" help:"uninstall topaz container"`
	Update     topaz.UpdateCmd          `cmd:"" help:"update topaz container version"`
	Version    VersionCmd               `cmd:"" help:"version information"`
	NoCheck    bool                     `flag:"" name:"no-check" json:"no_check,omitempty" short:"N" env:"TOPAZ_NO_CHECK" help:"disable local container status check"`
	NoColor    bool                     `flag:"" name:"no-color" json:"no_color,omitempty" env:"TOPAZ_NO_COLOR" help:"disable colored terminal output"`
	LogLevel   int                      `flag:"" name:"verbosity" short:"v" type:"counter" default:"0" help:"log level"`
}
