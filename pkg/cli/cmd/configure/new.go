package configure

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/certs"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/common"
	"github.com/pkg/errors"
)

const (
	FromRemote = "remote"
	FromLocal  = "local"
)

type NewConfigCmd struct {
	Name             ConfigName `short:"n" help:"config name"`
	Resource         string     `short:"r" help:"policy uri (e.g. ghcr.io/org/policy:tag)"`
	From             string     `enum:"remote,local" default:"remote" help:"load policy from remote or local image"`
	Stdout           bool       `short:"p" help:"print to stdout" default:"false"`
	EdgeDirectory    bool       `short:"d" help:"enable edge directory" default:"false"`
	Force            bool       `short:"f" flag:"" default:"false" required:"false" help:"skip confirmation prompt"`
	LocalPolicyImage string     `short:"l" help:"[deprecated: use --local instead] local policy image name"`
}

func (cmd *NewConfigCmd) Run(c *cc.CommonCtx) error {
	if cmd.Resource == "" {
		if cmd.LocalPolicyImage == "" {
			return errors.New("no policy specified. Please provide a policy URI with the --resource (-r) option")
		} else {
			c.Con().Warn().Msg("The --local-policy-image options (-l) is deprecated and will be removed in a future release. " +
				"Please use the --local flag instead.")
		}
	}

	configFile := cmd.Name.String() + ".yaml"
	if configFile != c.Config.Active.ConfigFile {
		c.Config.Active.Config = cmd.Name.String()
		c.Config.Active.ConfigFile = filepath.Join(cc.GetTopazCfgDir(), configFile)
	}

	if !cmd.Stdout {
		c.Con().Info().Msg(">>> configure policy\n")
	}

	// Backward-compatibility with deprecated LocalPolicyImage option.
	resource, local := cmd.Resource, cmd.From == FromLocal
	if cmd.LocalPolicyImage != "" {
		resource, local = cmd.LocalPolicyImage, true
	}

	configGenerator := config.NewGenerator(cmd.Name.String()).
		WithVersion(config.ConfigFileVersion).
		WithPolicyName(cmd.Name.String()).
		WithResource(resource).
		WithLocalPolicy(local).
		WithEdgeDirectory(cmd.EdgeDirectory)

	_, err := configGenerator.CreateConfigDir()
	if err != nil {
		return err
	}

	if _, err := configGenerator.CreateCertsDir(); err != nil {
		return err
	}

	certGenerator := certs.GenerateCertsCmd{CertsDir: cc.GetTopazCertsDir()}
	err = certGenerator.Run(c)
	if err != nil {
		return err
	}

	if _, err := configGenerator.CreateDataDir(); err != nil {
		return err
	}

	var w io.Writer

	if cmd.Stdout {
		w = c.StdOut()
	} else {
		if !cmd.Force {
			if _, err := os.Stat(c.Config.Active.ConfigFile); err == nil {
				c.Con().Warn().Msg("Configuration file %q already exists.", c.Config.Active.ConfigFile)
				if !common.PromptYesNo("Do you want to continue?", false) {
					return nil
				}
			}
		}
		w, err = os.Create(c.Config.Active.ConfigFile)
		if err != nil {
			return err
		}
	}

	if !cmd.Stdout {
		if local {
			c.Con().Info().Msg("using local policy image: %s", resource)
			return configGenerator.GenerateConfig(w, config.LocalImageTemplate)
		}

		c.Con().Info().Msg("policy name: %s", cmd.Name)
	}

	c.Context = context.WithValue(c.Context, common.Save, true)

	return configGenerator.GenerateConfig(w, config.Template)
}
