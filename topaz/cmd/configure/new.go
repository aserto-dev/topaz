package configure

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/topaz/cc"
	"github.com/aserto-dev/topaz/topaz/cmd/certs"
	"github.com/aserto-dev/topaz/topaz/cmd/common"
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

//nolint:funlen,nestif
func (cmd *NewConfigCmd) Run(ctx context.Context) error {
	if cmd.Resource == "" {
		if cmd.LocalPolicyImage == "" {
			return errors.New("no policy specified. Please provide a policy URI with the --resource (-r) option")
		} else {
			cc.Con().Warn().Msg("The --local-policy-image options (-l) is deprecated and will be removed in a future release. " +
				"Please use the --local flag instead.")
		}
	}

	cfg := cc.GetConfig()

	configFile := cmd.Name.String() + ".yaml"
	if configFile != cfg.Active.ConfigFile {
		cfg.Active.Config = cmd.Name.String()
		cfg.Active.ConfigFile = filepath.Join(cc.GetTopazCfgDir(), configFile)
	}

	if !cmd.Stdout {
		cc.Con().Info().Msg(">>> configure policy\n")
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

	if _, err := configGenerator.CreateConfigDir(); err != nil {
		return err
	}

	if _, err := configGenerator.CreateCertsDir(); err != nil {
		return err
	}

	certGenerator := certs.GenerateCertsCmd{CertsDir: cc.GetTopazCertsDir()}
	if err := certGenerator.Run(ctx); err != nil {
		return err
	}

	if _, err := configGenerator.CreateDataDir(); err != nil {
		return err
	}

	var (
		w   io.Writer
		err error
	)

	if cmd.Stdout {
		w = os.Stdout
	} else {
		if !cmd.Force {
			if _, err := os.Stat(cfg.Active.ConfigFile); err == nil {
				cc.Con().Warn().Msg("Configuration file %q already exists.", cfg.Active.ConfigFile)

				if !common.PromptYesNo("Do you want to continue?", false) {
					return nil
				}
			}
		}

		w, err = os.Create(cfg.Active.ConfigFile)
		if err != nil {
			return err
		}
	}

	if !cmd.Stdout {
		if local {
			cc.Con().Info().Msg("using local policy image: %s", resource)
			return configGenerator.GenerateConfig(w, config.LocalImageTemplate)
		}

		cc.Con().Info().Msg("policy name: %s", cmd.Name)
	}

	return configGenerator.GenerateConfig(w, config.Template)
}
