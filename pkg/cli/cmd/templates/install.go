package templates

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	client "github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	azc "github.com/aserto-dev/topaz/pkg/cli/clients/authorizer"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/common"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/configure"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/directory"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/topaz"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type InstallTemplateCmd struct {
	Name              string `arg:"" required:"" help:"template name"`
	Force             bool   `flag:"" short:"f" default:"false" required:"false" help:"skip confirmation prompt"`
	NoConfigure       bool   `optional:"" default:"false" help:"do not run configure step, to prevent changes to the config .yaml file"`
	NoTests           bool   `optional:"" default:"false" help:"do not execute assertions as part of template installation"`
	NoConsole         bool   `optional:"" default:"false" help:"do not open console when template installation is finished"`
	ContainerRegistry string `optional:"" default:"${container_registry}" env:"CONTAINER_REGISTRY" help:"container registry (host[:port]/repo)"`
	ContainerImage    string `optional:"" default:"${container_image}" env:"CONTAINER_IMAGE" help:"container image name"`
	ContainerTag      string `optional:"" default:"${container_tag}" env:"CONTAINER_TAG" help:"container tag"`
	ContainerPlatform string `optional:"" default:"${container_platform}" env:"CONTAINER_PLATFORM" help:"container platform"`
	ContainerName     string `optional:"" default:"${container_name}" env:"CONTAINER_NAME" help:"container name"`
	ContainerHostname string `optional:"" name:"hostname" default:"" env:"CONTAINER_HOSTNAME" help:"hostname for docker to set"`
	TemplatesURL      string `arg:"" required:"false" default:"https://topaz.sh/assets/templates/templates.json" help:"URL of template catalog"`
	ContainerVersion  string `optional:"" hidden:"" default:"" env:"CONTAINER_VERSION"`
	ConfigName        string `optional:"" help:"set config name"`
	dsc.Config
}

func (cmd *InstallTemplateCmd) Run(c *cc.CommonCtx) error {
	cmd.ContainerTag = cc.ContainerVersionTag(cmd.ContainerVersion, cmd.ContainerTag)

	tmpl, err := getTemplate(cmd.Name, cmd.TemplatesURL)
	if err != nil {
		return err
	}

	if !cmd.Force {
		fmt.Fprintln(c.StdErr(), "Installing this template will completely reset your topaz configuration.")
		if !common.PromptYesNo("Do you want to continue?", false) {
			return nil
		}
	}
	fileName := fmt.Sprintf("%s.yaml", tmpl.Name)
	c.Config.Active.Config = tmpl.Name
	if cmd.ConfigName != "" {
		if !common.RestrictedNamePattern.MatchString(cmd.ConfigName) {
			return errors.Errorf("%s must match pattern %s", cmd.Name, common.RestrictedNamePattern.String())
		}
		fileName = fmt.Sprintf("%s.yaml", cmd.ConfigName)
		c.Config.Active.Config = cmd.ConfigName
	}

	// reset defaults on template install
	c.Config.Active.ConfigFile = filepath.Join(cc.GetTopazCfgDir(), fileName)
	c.Config.Running.ActiveConfig = c.Config.Active
	c.Config.Running.ContainerName = cc.ContainerName(c.Config.Active.ConfigFile)
	cmd.ContainerName = c.Config.Running.ContainerName

	if _, err := os.Stat(cc.GetTopazDir()); os.IsNotExist(err) {
		err = os.MkdirAll(cc.GetTopazDir(), 0o700)
		if err != nil {
			return err
		}
	}

	cliConfig := filepath.Join(cc.GetTopazDir(), common.CLIConfigurationFile)

	kongConfigBytes, err := json.Marshal(c.Config)
	if err != nil {
		return err
	}
	err = os.WriteFile(cliConfig, kongConfigBytes, 0o666) // nolint
	if err != nil {
		return err
	}

	return cmd.installTemplate(c, tmpl)
}

// installTemplate steps:
// 1 - topaz stop - ensure topaz is not running, so we can reconfigure
// 2 - topaz config new - generate a new configuration based on the requirements of the template
// 3 - topaz start - start instance using new configuration
// 4 - wait for health endpoint to be in serving state
// 5 - topaz manifest delete --force, reset the directory store
// 6 - topaz manifest set, deploy the manifest
// 7 - topaz import, load IDP and domain data (in that order)
// 8 - topaz test exec, execute assertions when part of template
// 9 - topaz console, launch console so the user start exploring the template artifacts.
func (cmd *InstallTemplateCmd) installTemplate(c *cc.CommonCtx, tmpl *template) error {
	topazTemplateDir := cc.GetTopazTemplateDir()

	cmd.Config.Insecure = true
	// 1-3 - stop topaz, configure, start
	if err := cmd.prepareTopaz(c, tmpl, cmd.ConfigName); err != nil {
		return err
	}

	// 4 - wait for health endpoint to be in serving state
	cfg := config.GetConfig(c.Config.Active.ConfigFile)
	if cfg.HasTopazDir {
		fmt.Fprintln(c.StdErr(), "This configuration file still uses TOPAZ_DIR environment variable.")
		fmt.Fprintln(c.StdErr(), "Please change to using the new TOPAZ_DB_DIR and TOPAZ_CERTS_DIR environment variables.")
	}

	healthCfg := &client.Config{
		Address:        cfg.Configuration.APIConfig.Health.ListenAddress,
		ClientCertPath: "",
		ClientKeyPath:  "",
		CACertPath:     "",
		Insecure:       false,
		NoTLS:          true,
		NoProxy:        false,
	}

	if healthy, err := cc.ServiceHealthStatus(c.Context, healthCfg, "model"); err != nil {
		return errors.Wrapf(err, "unable to check health status")
	} else if !healthy {
		return err
	}

	if model, ok := cfg.Configuration.APIConfig.Services["model"]; !ok {
		return errors.Errorf("model service not configured")
	} else {
		cmd.Config.Host = model.GRPC.ListenAddress
	}

	// 5-7 - reset directory, apply (manifest, IDP and domain data) template.
	if err := installTemplate(c, tmpl, topazTemplateDir, &cmd.Config, cmd.ConfigName).Install(); err != nil {
		return err
	}

	// 8 - run tests
	if !cmd.NoTests {
		if err := installTemplate(c, tmpl, topazTemplateDir, &cmd.Config, cmd.ConfigName).Test(); err != nil {
			return err
		}
	}

	// 9 - topaz console, launch console so the user start exploring the template artifacts
	if !cmd.NoConsole {
		command := topaz.ConsoleCmd{
			ConsoleAddress: "https://localhost:8080/ui/directory",
		}
		if err := command.Run(c); err != nil {
			return err
		}
	}

	return nil
}

func (cmd *InstallTemplateCmd) prepareTopaz(c *cc.CommonCtx, tmpl *template, customName string) error {
	// 1 - topaz stop - ensure topaz is not running, so we can reconfigure
	{
		command := &topaz.StopCmd{
			ContainerName: "topaz*",
			Wait:          true,
		}
		if err := command.Run(c); err != nil {
			return err
		}
	}

	// topaz status, output status
	{
		command := &topaz.StatusCmd{}
		if err := command.Run(c); err != nil {
			return err
		}
	}

	name := tmpl.Assets.Policy.Name
	if customName != "" {
		name = customName
	}

	// 2 - topaz config new - generate a new configuration based on the requirements of the template
	if !cmd.NoConfigure {
		command := configure.NewConfigCmd{
			Name:     configure.ConfigName(name),
			Resource: tmpl.Assets.Policy.Resource,
			From:     lo.Ternary(tmpl.Assets.Policy.Local, configure.FromLocal, configure.FromRemote),
			Force:    true,
		}
		if err := command.Run(c); err != nil {
			return err
		}
	}

	// topaz config use - activate configuration (new or existing)
	{
		use := configure.UseConfigCmd{
			Name:      configure.ConfigName(name),
			ConfigDir: cc.GetTopazCfgDir(),
		}
		if err := use.Run(c); err != nil {
			return err
		}
	}

	// 3 - topaz start - start instance using new configuration
	{
		command := &topaz.StartCmd{
			StartRunCmd: topaz.StartRunCmd{
				ContainerRegistry: cmd.ContainerRegistry,
				ContainerImage:    cmd.ContainerImage,
				ContainerTag:      cmd.ContainerTag,
				ContainerPlatform: cmd.ContainerPlatform,
				ContainerName:     cmd.ContainerName,
				ContainerHostname: cmd.ContainerHostname,
			},
			Wait: true,
		}
		if err := command.Run(c); err != nil {
			return err
		}
	}

	return nil
}

func installTemplate(c *cc.CommonCtx, tmpl *template, topazTemplateDir string, cfg *dsc.Config, customName string) *tmplInstaller {
	return &tmplInstaller{
		c:                c,
		tmpl:             tmpl,
		topazTemplateDir: topazTemplateDir,
		cfg:              cfg,
		customName:       customName,
	}
}

type tmplInstaller struct {
	c                *cc.CommonCtx
	tmpl             *template
	topazTemplateDir string
	cfg              *dsc.Config
	customName       string
}

func (i *tmplInstaller) Install() error {
	// 5 - topaz manifest delete --force, reset the directory store
	if err := i.deleteManifest(); err != nil {
		return err
	}

	// 6 - topaz manifest set, apply the manifest
	if err := i.setManifest(); err != nil {
		return err
	}

	// 7 - topaz import, load IDP and domain data
	if err := i.importData(); err != nil {
		return err
	}

	return nil
}

func (i *tmplInstaller) Test() error {
	// 8 - topaz test exec, execute assertions when part of template
	return i.runTemplateTests()
}

func (i *tmplInstaller) deleteManifest() error {
	command := directory.DeleteManifestCmd{
		Force:  true,
		Config: *i.cfg,
	}
	return command.Run(i.c)
}

func (i *tmplInstaller) setManifest() error {
	manifest := i.tmpl.AbsURL(i.tmpl.Assets.Manifest)

	name := i.tmpl.Name
	if i.customName != "" {
		name = i.customName
	}

	if exists, _ := config.FileExists(manifest); !exists {
		manifestDir := path.Join(i.topazTemplateDir, name, "model")
		switch m, err := download(manifest, manifestDir); {
		case err != nil:
			return err
		default:
			manifest = m
		}
	}

	command := directory.SetManifestCmd{
		Path:   manifest,
		Config: *i.cfg,
	}

	return command.Run(i.c)
}

func (i *tmplInstaller) importData() error {
	name := i.tmpl.Name
	if i.customName != "" {
		name = i.customName
	}

	defaultDataDir := path.Join(i.topazTemplateDir, name, "data")

	dataDirs := map[string]struct{}{}
	for _, v := range append(i.tmpl.Assets.IdentityData, i.tmpl.Assets.DomainData...) {
		dataURL := i.tmpl.AbsURL(v)
		if exists, _ := config.FileExists(dataURL); exists {
			dataDirs[path.Dir(dataURL)] = struct{}{}
			continue
		}

		if _, err := download(dataURL, defaultDataDir); err != nil {
			return err
		}
		dataDirs[defaultDataDir] = struct{}{}
	}

	for dir := range dataDirs {
		command := directory.ImportCmd{
			Directory: dir,
			Config:    *i.cfg,
		}

		if err := command.Run(i.c); err != nil {
			return err
		}
	}

	return nil
}

func (i *tmplInstaller) runTemplateTests() error {
	name := i.tmpl.Name
	if i.customName != "" {
		name = i.customName
	}
	assertionsDir := path.Join(i.topazTemplateDir, name, "assertions")

	tests := []string{}
	for _, v := range i.tmpl.Assets.Assertions {
		assertionURL := i.tmpl.AbsURL(v)
		if exists, _ := config.FileExists(assertionURL); exists {
			tests = append(tests, assertionURL)
			continue
		}
		switch t, err := download(assertionURL, assertionsDir); {
		case err != nil:
			return err
		default:
			tests = append(tests, t)
		}
	}

	runner, err := common.NewTestRunner(
		i.c,
		&common.TestExecCmd{
			Files:   tests,
			Stdin:   false,
			Summary: true,
			Format:  "table",
			Desc:    "on-error",
		},
		&azc.Config{
			Host:      cc.AuthorizerSvc(),
			APIKey:    i.cfg.APIKey,
			Token:     i.cfg.Token,
			Insecure:  i.cfg.Insecure,
			Plaintext: i.cfg.Plaintext,
			TenantID:  i.cfg.TenantID,
			Headers:   i.cfg.Headers,
			Timeout:   i.cfg.Timeout,
		},
		&dsc.Config{
			Host:      i.cfg.Host,
			APIKey:    i.cfg.APIKey,
			Token:     i.cfg.Token,
			Insecure:  i.cfg.Insecure,
			Plaintext: i.cfg.Plaintext,
			TenantID:  i.cfg.TenantID,
			Headers:   i.cfg.Headers,
			Timeout:   i.cfg.Timeout,
		},
	)
	if err != nil {
		return err
	}

	if err := runner.Run(i.c); err != nil {
		return err
	}

	return nil
}
