package configure

import (
	"io"
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/certs"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/common"
	"github.com/aserto-dev/topaz/pkg/cli/config"
)

const (
	FromRemote = "remote"
	FromLocal  = "local"
)

type NewConfigCmd struct {
	Name          ConfigName `short:"n" help:"config name"`
	Resource      string     `short:"r" required:"true" help:"policy uri or path (e.g. ghcr.io/org/policy:tag)"`
	From          string     `enum:"remote,local" default:"remote" help:"load policy from remote or local image"`
	Stdout        bool       `short:"p" help:"print to stdout" default:"false"`
	EdgeDirectory bool       `short:"d" help:"enable edge directory" default:"false"`
	Force         bool       `short:"f" flag:"" default:"false" required:"false" help:"skip confirmation prompt"`
}

func (cmd *NewConfigCmd) Run(c *cc.CommonCtx) error {
	configFile := cmd.Name.String() + ".yaml"
	if configFile != c.Config.Active.ConfigFile {
		c.Config.Active.Config = cmd.Name.String()
		c.Config.Active.ConfigFile = filepath.Join(cc.GetTopazCfgDir(), configFile)
	}

	if !cmd.Stdout {
		c.Con().Info().Msg(">>> configure policy\n")
	}

	certGenerator := certs.GenerateCertsCmd{CertsDir: cc.GetTopazCertsDir()}
	if err := certGenerator.Run(c); err != nil {
		return err
	}

	w, err := cmd.writer(c)
	if err != nil {
		return err
	}

	c.Con().Info().Msg("policy name: %s", cmd.Name)

	gen, err := config.NewGenerator(cmd.Name.String())
	if err != nil {
		return err
	}

	local := cmd.From == FromLocal
	if local {
		gen.WithLocalBundle(cmd.Resource)
		c.Con().Info().Msg("using local policy: %s", cmd.Resource)
	} else {
		gen.WithBundle(cmd.Name.String(), cmd.Resource)
	}

	return gen.Generate(w)
}

func (cmd *NewConfigCmd) writer(c *cc.CommonCtx) (io.Writer, error) {
	if cmd.Stdout {
		return c.StdOut(), nil
	}

	if !cmd.Force {
		if _, err := os.Stat(c.Config.Active.ConfigFile); err == nil {
			c.Con().Warn().Msg("Configuration file %q already exists.", c.Config.Active.ConfigFile)

			if !common.PromptYesNo("Do you want to continue?", false) {
				return nil, nil
			}
		}
	}

	return os.Create(c.Config.Active.ConfigFile)
}
