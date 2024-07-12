package configure

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/certs"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/common"
	"github.com/fatih/color"
	"github.com/pkg/errors"
)

type NewConfigCmd struct {
	Name             ConfigName `short:"n" help:"config name"`
	LocalPolicyImage string     `short:"l" help:"local policy image name"`
	Resource         string     `short:"r" help:"resource url"`
	Stdout           bool       `short:"p" help:"print to stdout"`
	EdgeDirectory    bool       `short:"d" help:"enable edge directory" default:"false"`
	Force            bool       `flag:"" default:"false" short:"f" required:"false" help:"skip confirmation prompt"`
}

func (cmd *NewConfigCmd) Run(c *cc.CommonCtx) error {
	if cmd.Name == "" && cmd.Resource == "" {
		if cmd.LocalPolicyImage == "" {
			return errors.New("you either need to provide a local policy image or the resource and the policy name for the configuration")
		}
	}

	configFile := cmd.Name.String() + ".yaml"
	if configFile != c.Config.Active.ConfigFile {
		c.Config.Active.Config = cmd.Name.String()
		c.Config.Active.ConfigFile = filepath.Join(cc.GetTopazCfgDir(), configFile)
	}

	if !cmd.Stdout {
		_, _ = fmt.Fprint(c.StdErr(), color.GreenString(">>> configure policy\n"))
	}

	configGenerator := config.NewGenerator(cmd.Name.String()).
		WithVersion(config.ConfigFileVersion).
		WithLocalPolicyImage(cmd.LocalPolicyImage).
		WithPolicyName(cmd.Name.String()).
		WithResource(cmd.Resource).
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
				fmt.Fprintf(c.StdErr(), "Configuration file %q already exists.\n", c.Config.Active.ConfigFile)
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
		if cmd.LocalPolicyImage != "" {
			fmt.Fprint(c.StdOut(), color.GreenString("using local policy image: %s\n", cmd.LocalPolicyImage))
			return configGenerator.GenerateConfig(w, config.LocalImageTemplate)
		}

		fmt.Fprint(c.StdOut(), color.GreenString("policy name: %s\n", cmd.Name))
	}

	c.Context = context.WithValue(c.Context, common.Save, true)

	return configGenerator.GenerateConfig(w, config.Template)
}
