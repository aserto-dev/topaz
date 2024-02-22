package cmd

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/fatih/color"
)

type ConfigureCmd struct {
	PolicyName        string `short:"n" help:"policy name"`
	LocalPolicyImage  string `short:"l" help:"local policy image name"`
	Resource          string `short:"r" help:"resource url"`
	Stdout            bool   `short:"p" help:"print to stdout"`
	EdgeDirectory     bool   `short:"d" help:"enable edge directory" default:"false"`
	Force             bool   `flag:"" default:"false" short:"f" required:"false" help:"skip confirmation prompt"`
	EnableDirectoryV2 bool   `flag:"" name:"enable-v2" hidden:"" default:"true" help:"enable directory version 2 services for backwards compatibility"`
	ConfigFile        string `name:"config" json:"config,omitempty" default:"config.yaml" short:"c" help:"topaz configuration file"`
}

func (cmd *ConfigureCmd) Run(c *cc.CommonCtx) error {
	if cmd.PolicyName == "" && cmd.Resource == "" {
		if cmd.LocalPolicyImage == "" {
			return errors.New("you either need to provide a local policy image or the resource and the policy name for the configuration")
		}
	}
	if cmd.ConfigFile != c.Config.DefaultConfigFile {
		c.Config.DefaultConfigFile = filepath.Join(cc.GetTopazCfgDir(), cmd.ConfigFile)
		c.Config.ContainerName = cc.ContainerName(c.Config.DefaultConfigFile)
	}
	if !cmd.Stdout {
		color.Green(">>> configure policy")
	}

	configGenerator := config.NewGenerator(cmd.PolicyName).
		WithVersion(config.ConfigFileVersion).
		WithLocalPolicyImage(cmd.LocalPolicyImage).
		WithPolicyName(cmd.PolicyName).
		WithResource(cmd.Resource).
		WithEdgeDirectory(cmd.EdgeDirectory).
		WithEnableDirectoryV2(cmd.EnableDirectoryV2)

	_, err := configGenerator.CreateConfigDir()
	if err != nil {
		return err
	}

	if _, err := configGenerator.CreateCertsDir(); err != nil {
		return err
	}

	certGenerator := GenerateCertsCmd{CertsDir: cc.GetTopazCertsDir()}
	err = certGenerator.Run(c)
	if err != nil {
		return err
	}

	if _, err := configGenerator.CreateDataDir(); err != nil {
		return err
	}

	var w io.Writer

	if cmd.Stdout {
		w = c.UI.Output()
	} else {
		if !cmd.Force {
			if _, err := os.Stat(c.Config.DefaultConfigFile); err == nil {
				c.UI.Exclamation().Msg("A configuration file already exists.")
				if !promptYesNo("Do you want to continue?", false) {
					return nil
				}
			}
		}
		w, err = os.Create(c.Config.DefaultConfigFile)
		if err != nil {
			return err
		}
	}

	if !cmd.Stdout {
		if cmd.LocalPolicyImage != "" {
			color.Green("using local policy image: %s", cmd.LocalPolicyImage)
			return configGenerator.GenerateConfig(w, config.LocalImageTemplate)
		}

		color.Green("policy name: %s", cmd.PolicyName)
	}

	return configGenerator.GenerateConfig(w, config.Template)
}
