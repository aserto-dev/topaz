package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/adrg/xdg"
	"github.com/alecthomas/kong"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/fatih/color"
	"github.com/pkg/errors"
)

type ConfigCmd struct {
	Use    UseConfigCmd    `cmd:"" help:"set active configuration"`
	New    NewConfigCmd    `cmd:"" help:"create new configuration"`
	List   ListConfigCmd   `cmd:"" help:"list configurations"`
	Rename RenameConfigCmd `cmd:"" help:"rename configuration"`
	Delete DeleteConfigCmd `cmd:"" help:"delete configuration"`
	Info   InfoConfigCmd   `cmd:"" help:"display configuration information"`
}

var restrictedNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]*$`)

type ConfigName string

func (c ConfigName) AfterApply(ctx *kong.Context) error {
	if string(c) == "" {
		return fmt.Errorf("no configuration name value provided")
	}

	if !restrictedNamePattern.MatchString(string(c)) {
		return fmt.Errorf("configuration name is invalid, must match pattern %q", restrictedNamePattern.String())
	}

	return nil
}

func (c ConfigName) String() string {
	return string(c)
}

func (cmd *ConfigCmd) Run(c *cc.CommonCtx) error {
	return nil
}

type DeleteConfigCmd struct {
	Name      ConfigName `arg:"" required:"" help:"topaz config name"`
	ConfigDir string     `flag:"" required:"false" default:"${topaz_cfg_dir}" help:"path to config folder" `
}

func (cmd *DeleteConfigCmd) Run(c *cc.CommonCtx) error {
	if c.CheckRunStatus(cc.ContainerName(fmt.Sprintf("%s.yaml", cmd.Name)), cc.StatusRunning) {
		return fmt.Errorf("configuration %q is running, use 'topaz stop' to stop, before deleting", cmd.Name)
	}

	c.UI.Normal().Msgf("Removing configuration %q", cmd.Name)
	filename := filepath.Join(cmd.ConfigDir, fmt.Sprintf("%s.yaml", cmd.Name))

	if c.Config.Active.Config == cmd.Name.String() {
		c.Config.Active.Config = ""
		c.Config.Active.ConfigFile = ""
		if err := c.SaveContextConfig(CLIConfigurationFile); err != nil {
			return errors.Wrap(err, "failed to update active context")
		}
	}

	return os.Remove(filename)

}

type RenameConfigCmd struct {
	Name      ConfigName `arg:"" required:"" help:"topaz config name"`
	NewName   ConfigName `arg:"" required:"" help:"topaz new config name"`
	ConfigDir string     `flag:"" required:"false" default:"${topaz_cfg_dir}" help:"path to config folder" `
}

func (cmd *RenameConfigCmd) Run(c *cc.CommonCtx) error {
	if c.CheckRunStatus(cc.ContainerName(fmt.Sprintf("%s.yaml", cmd.Name)), cc.StatusRunning) {
		return fmt.Errorf("configuration %q is running, use 'topaz stop' to stop, before renaming", cmd.Name)
	}

	c.UI.Normal().Msgf("Renaming configuration %q to %q", cmd.Name, cmd.NewName)
	oldFile := filepath.Join(cmd.ConfigDir, fmt.Sprintf("%s.yaml", cmd.Name))
	newFile := filepath.Join(cmd.ConfigDir, fmt.Sprintf("%s.yaml", cmd.NewName))

	if c.Config.Active.Config == cmd.Name.String() {
		c.Config.Active.Config = cmd.NewName.String()
		c.Config.Active.ConfigFile = newFile
		if err := c.SaveContextConfig(CLIConfigurationFile); err != nil {
			return errors.Wrap(err, "failed to update active context")
		}
	}

	return os.Rename(oldFile, newFile)
}

type UseConfigCmd struct {
	Name      ConfigName `arg:"" required:"" help:"topaz config name"`
	ConfigDir string     `flag:"" required:"false" default:"${topaz_cfg_dir}" help:"path to config folder" `
}

func (cmd *UseConfigCmd) Run(c *cc.CommonCtx) error {
	if _, err := os.Stat(filepath.Join(cmd.ConfigDir, fmt.Sprintf("%s.yaml", cmd.Name))); err != nil {
		return err
	}

	topazContainers, err := c.GetRunningContainers()
	if err != nil {
		return err
	}

	if len(topazContainers) > 0 {
		return cc.ErrIsRunning
	}

	c.Config.Active.Config = cmd.Name.String()
	c.Config.Active.ConfigFile = filepath.Join(cmd.ConfigDir, fmt.Sprintf("%s.yaml", cmd.Name))
	c.UI.Normal().Msgf("Using configuration %q", cmd.Name)

	if err := c.SaveContextConfig(CLIConfigurationFile); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	return nil
}

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
		_, _ = fmt.Fprint(color.Error, color.GreenString(">>> configure policy"))
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
			if _, err := os.Stat(c.Config.Active.ConfigFile); err == nil {
				c.UI.Exclamation().Msg("A configuration file already exists.")
				if !PromptYesNo("Do you want to continue?", false) {
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
		if files[i].Name() == filepath.Base(c.Config.Active.ConfigFile) {
			active = "*"
		}

		table.WithTableRow(active, name, files[i].Name())
	}
	table.Do()

	return nil
}

type InfoConfigCmd struct{}

func (cmd InfoConfigCmd) Run(c *cc.CommonCtx) error {
	data := map[string]interface{}{
		// environment values
		"environment": map[string]interface{}{
			"HOME":            xdg.Home,
			"XDG_CONFIG_HOME": xdg.ConfigHome,
			"XDG_DATA_HOME":   xdg.DataHome,
		},
		// config values
		"config": map[string]interface{}{
			"TOPAZ_CFG_DIR":   cc.GetTopazCfgDir(),
			"TOPAZ_CERTS_DIR": cc.GetTopazCertsDir(),
			"TOPAZ_DB_DIR":    cc.GetTopazDataDir(),
			"TOPAZ_DIR":       cc.GetTopazDir(),
		},
		// runtime values
		"runtime": map[string]interface{}{
			"active.configuration.name":  c.Config.Active.Config,
			"active.configuration.file":  c.Config.Active.ConfigFile,
			"running.configuration.name": c.Config.Running.Config,
			"running.configuration.file": c.Config.Running.ConfigFile,
			"running.container.name":     c.Config.Running.ContainerName,
			"topaz.json":                 filepath.Join(cc.GetTopazDir(), CLIConfigurationFile),
		},
		// default values
		"default": map[string]interface{}{
			"CONTAINER_IMAGE":    c.Config.Defaults.ContainerImage,
			"CONTAINER_TAG":      c.Config.Defaults.ContainerTag,
			"CONTAINER_PLATFORM": c.Config.Defaults.ContainerPlatform,
			"TOPAZ_NO_CHECK":     strconv.FormatBool(c.Config.Defaults.NoCheck),
		},
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	enc.Encode(data)

	return nil
}
