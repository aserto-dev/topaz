package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/fatih/color"
)

type ConfigCmd struct {
	Use    UseConfigCmd    `cmd:"" help:"set active configuration"`
	New    NewConfigCmd    `cmd:"" help:"create new configuration"`
	List   ListConfigCmd   `cmd:"" help:"list configurations"`
	Rename RenameConfigCmd `cmd:"" help:"rename configuration"`
	Delete DeleteConfigCmd `cmd:"" help:"delete configuration"`
}

func (cmd *ConfigCmd) Run(c *cc.CommonCtx) error {
	return nil
}

type DeleteConfigCmd struct {
	ConfigDir string `flag:"" required:"false" default:"${topaz_cfg_dir}" help:"path to config folder" `
	Name      string `arg:"" required:"" help:"topaz config name"`
}

func (cmd *DeleteConfigCmd) Run(c *cc.CommonCtx) error {
	if cmd.Name == "" {
		return errors.New("configuration name must be provided")
	}
	if c.CheckRunStatus(cc.ContainerName(fmt.Sprintf("%s.yaml", cmd.Name)), cc.StatusRunning) {
		return cc.ErrIsRunning
	}
	c.UI.Normal().Msgf("Removing configuration %s", fmt.Sprintf("%s.yaml", cmd.Name))
	filename := filepath.Join(cmd.ConfigDir, fmt.Sprintf("%s.yaml", cmd.Name))

	return os.Remove(filename)
}

type RenameConfigCmd struct {
	ConfigDir string `flag:"" required:"false" default:"${topaz_cfg_dir}" help:"path to config folder" `
	Name      string `arg:"" required:"" help:"topaz config name"`
	NewName   string `arg:"" required:"" help:"topaz new config name"`
}

func (cmd *RenameConfigCmd) Run(c *cc.CommonCtx) error {
	if cmd.Name == "" || cmd.NewName == "" {
		return errors.New("old configuration name and new configuration name must be provided")
	}
	if c.CheckRunStatus(cc.ContainerName(fmt.Sprintf("%s.yaml", cmd.Name)), cc.StatusRunning) {
		return cc.ErrIsRunning
	}
	c.UI.Normal().Msgf("Renamig configuration %s to %s", fmt.Sprintf("%s.yaml", cmd.Name), fmt.Sprintf("%s.yaml", cmd.NewName))
	oldFile := filepath.Join(cmd.ConfigDir, fmt.Sprintf("%s.yaml", cmd.Name))
	newFile := filepath.Join(cmd.ConfigDir, fmt.Sprintf("%s.yaml", cmd.NewName))

	return os.Rename(oldFile, newFile)
}

type UseConfigCmd struct {
	ConfigDir string `flag:"" required:"false" default:"${topaz_cfg_dir}" help:"path to config folder" `
	Name      string `arg:"" required:"" help:"topaz config name"`
}

func (cmd *UseConfigCmd) Run(c *cc.CommonCtx) error {
	if cmd.Name == "" {
		return errors.New("configuration name must be provided")
	}
	if _, err := os.Stat(filepath.Join(cmd.ConfigDir, fmt.Sprintf("%s.yaml", cmd.Name))); err != nil {
		return err
	}

	c.Config.TopazConfigFile = filepath.Join(cmd.ConfigDir, fmt.Sprintf("%s.yaml", cmd.Name))
	c.Config.ContainerName = cc.ContainerName(c.Config.TopazConfigFile)
	c.UI.Normal().Msgf("Using configuration %s", fmt.Sprintf("%s.yaml", cmd.Name))

	if err := c.SaveContextConfig(CLIConfigurationFile); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	return nil
}

type NewConfigCmd struct {
	Name              string `short:"n" help:"config name"`
	LocalPolicyImage  string `short:"l" help:"local policy image name"`
	Resource          string `short:"r" help:"resource url"`
	Stdout            bool   `short:"p" help:"print to stdout"`
	EdgeDirectory     bool   `short:"d" help:"enable edge directory" default:"false"`
	Force             bool   `flag:"" default:"false" short:"f" required:"false" help:"skip confirmation prompt"`
	EnableDirectoryV2 bool   `flag:"" name:"enable-v2" hidden:"" default:"true" help:"enable directory version 2 services for backwards compatibility"`
}

func (cmd *NewConfigCmd) Run(c *cc.CommonCtx) error {
	if cmd.Name == "" && cmd.Resource == "" {
		if cmd.LocalPolicyImage == "" {
			return errors.New("you either need to provide a local policy image or the resource and the policy name for the configuration")
		}
	}

	configFile := cmd.Name + ".yaml"
	if configFile != c.Config.TopazConfigFile {
		c.Config.TopazConfigFile = filepath.Join(cc.GetTopazCfgDir(), configFile)
		c.Config.ContainerName = cc.ContainerName(c.Config.TopazConfigFile)
	}

	if !cmd.Stdout {
		_, _ = fmt.Fprint(color.Error, color.GreenString(">>> configure policy"))
	}

	configGenerator := config.NewGenerator(cmd.Name).
		WithVersion(config.ConfigFileVersion).
		WithLocalPolicyImage(cmd.LocalPolicyImage).
		WithPolicyName(cmd.Name).
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
			if _, err := os.Stat(c.Config.TopazConfigFile); err == nil {
				c.UI.Exclamation().Msg("A configuration file already exists.")
				if !PromptYesNo("Do you want to continue?", false) {
					return nil
				}
			}
		}
		w, err = os.Create(c.Config.TopazConfigFile)
		if err != nil {
			return err
		}
	}

	if !cmd.Stdout {
		if cmd.LocalPolicyImage != "" {
			color.Green("using local policy image: %s", cmd.LocalPolicyImage)
			return configGenerator.GenerateConfig(w, config.LocalImageTemplate)
		}

		color.Green("policy name: %s", cmd.Name)
	}

	c.Context = context.WithValue(c.Context, Save, true)

	return configGenerator.GenerateConfig(w, config.Template)
}

type ListConfigCmd struct {
	ConfigDir string `flag:"" required:"false" default:"${topaz_cfg_dir}" help:"path to config folder" `
}

func (cmd ListConfigCmd) Run(c *cc.CommonCtx) error {
	table := c.UI.Normal().WithTable("", "Name", "Config File")
	files, err := os.ReadDir(cmd.ConfigDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for i := range files {
		name := strings.Split(files[i].Name(), ".")[0]
		active := ""
		if files[i].Name() == filepath.Base(c.Config.TopazConfigFile) {
			active = "*"
		}

		table.WithTableRow(active, name, files[i].Name())
	}
	table.Do()

	return nil
}
