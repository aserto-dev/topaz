package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	v3 "github.com/aserto-dev/azm/v3"
	"github.com/aserto-dev/go-directory/pkg/derr"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/rs/zerolog"
)

type TemplateCmd struct {
	List    ListTemplatesCmd   `cmd:"" help:"list template"`
	Install InstallTemplateCmd `cmd:"" help:"install template"`
	Verify  VerifyTemplateCmd  `cmd:"" help:"verify template content links" hidden:""`
}

type ListTemplatesCmd struct {
	TemplatesURL string `arg:"" required:"false" default:"https://topaz.sh/assets/templates/templates.json" help:"URL of template catalog"`
}

func (cmd *ListTemplatesCmd) Run(c *cc.CommonCtx) error {
	ctlg, err := getCatalog(cmd.TemplatesURL)
	if err != nil {
		return err
	}

	maxWidth := 0
	for n := range ctlg {
		maxWidth = max(maxWidth, len(n)+1)
	}

	table := c.UI.Normal().WithTable(colName, colDescription, colDocumentation)
	table.WithTableNoAutoWrapText()
	for n, t := range ctlg {
		table.WithTableRow(n, t.ShortDescription, t.DocumentationURL)
	}
	table.Do()

	return nil
}

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
	clients.Config
}

func (cmd *InstallTemplateCmd) Run(c *cc.CommonCtx) error {
	cmd.ContainerTag = cc.ContainerVersionTag(cmd.ContainerVersion, cmd.ContainerTag)

	tmpl, err := getTemplate(cmd.Name, cmd.TemplatesURL)
	if err != nil {
		return err
	}

	if !cmd.Force {
		c.UI.Exclamation().Msg("Installing this template will completely reset your topaz configuration.")
		if !promptYesNo("Do you want to continue?", false) {
			return nil
		}
	}

	// reset defaults on template install
	c.Config.DefaultConfigFile = filepath.Join(cc.GetTopazCfgDir(), fmt.Sprintf("%s.yaml", tmpl.Name))
	c.Config.ContainerName = cc.ContainerName(c.Config.DefaultConfigFile)

	cliConfig := filepath.Join(cc.GetTopazDir(), CLIConfigurationFile)

	kongConfigBytes, err := json.Marshal(c.Config)
	if err != nil {
		return err
	}
	err = os.WriteFile(cliConfig, kongConfigBytes, 0666) // nolint
	if err != nil {
		return err
	}

	return cmd.installTemplate(c, tmpl)
}

// installTemplate steps:
// 1 - topaz stop - ensure topaz is not running, so we can reconfigure
// 2 - topaz configure - generate a new configuration based on the requirements of the template
// 3 - topaz start - start instance using new configuration
// 4 - wait for health endpoint to be in serving state
// 5 - topaz manifest delete --force, reset the directory store
// 6 - topaz manifest set, deploy the manifest
// 7 - topaz import, load IDP and domain data (in that order)
// 8 - topaz test exec, execute assertions when part of template
// 9 - topaz console, launch console so the user start exploring the template artifacts.
func (cmd *InstallTemplateCmd) installTemplate(c *cc.CommonCtx, tmpl *template) error {
	topazDir, err := getTopazDir()
	if err != nil {
		return err
	}
	os.Setenv("TOPAZ_DIR", topazDir)

	cmd.Config.Insecure = true
	// 1-3 - stop topaz, configure, start
	err = cmd.prepareTopaz(c, tmpl)
	if err != nil {
		return err
	}

	// 4 - wait for health endpoint to be in serving state
	cfg := config.CurrentConfig()
	addr, _ := cfg.HealthService()
	if !cc.ServiceHealthStatus(addr, "") {
		return fmt.Errorf("gRPC endpoint not SERVING")
	}

	// 5-7 - reset directory, apply (manifest, IDP and domain data) template.
	if err := installTemplate(c, tmpl, topazDir, &cmd.Config).Install(); err != nil {
		return err
	}

	// 8 - run tests
	if !cmd.NoTests {
		if err := installTemplate(c, tmpl, topazDir, &cmd.Config).Test(); err != nil {
			return err
		}
	}

	// 9 - topaz console, launch console so the user start exploring the template artifacts
	if !cmd.NoConsole {
		command := ConsoleCmd{
			ConsoleAddress: "https://localhost:8080/ui/directory",
		}
		if err := command.Run(c); err != nil {
			return err
		}
	}

	return nil
}

func (cmd *InstallTemplateCmd) prepareTopaz(c *cc.CommonCtx, tmpl *template) error {

	// 1 - topaz stop - ensure topaz is not running, so we can reconfigure
	{
		command := &StopCmd{
			ContainerName: cmd.ContainerName,
			Wait:          true,
		}
		if err := command.Run(c); err != nil {
			return err
		}
	}

	// 2 - topaz configure - generate a new configuration based on the requirements of the template
	if !cmd.NoConfigure {
		command := ConfigureCmd{
			ConfigFile: fmt.Sprintf("%s.yaml", tmpl.Name),
			PolicyName: tmpl.Assets.Policy.Name,
			Resource:   tmpl.Assets.Policy.Resource,
			Force:      true,
			EnableDirectoryV2: false,
		}
		if err := command.Run(c); err != nil {
			return err
		}
	}
	// 3 - topaz start - start instance using new configuration
	{
		command := &StartCmd{
			StartRunCmd: StartRunCmd{
				ConfigFile:        fmt.Sprintf("%s.yaml", tmpl.Name),
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

const (
	colName          string = "name"
	colDescription   string = "description"
	colDocumentation string = "documentation"
)

type tmplCatalog map[string]*tmplRef

type tmplRef struct {
	Title            string `json:"title,omitempty"`
	ShortDescription string `json:"short_description,omitempty"`
	Description      string `json:"description,omitempty"`
	URL              string `json:"topaz_url,omitempty"`
	DocumentationURL string `json:"docs_url,omitempty"`

	absURL *url.URL
}

type template struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Assets      struct {
		Manifest string `json:"manifest"`
		Policy   struct {
			Name     string `json:"name"`
			Resource string `json:"resource"`
		} `json:"policy,omitempty"`
		IdentityData []string `json:"idp_data,omitempty"`
		DomainData   []string `json:"domain_data,omitempty"`
		Assertions   []string `json:"assertions,omitempty"`
	} `json:"assets"`

	base *url.URL
}

func (t *template) AbsURL(relative string) string {
	abs := *t.base
	abs.Path = path.Join(abs.Path, relative)
	return abs.String()
}

func installTemplate(c *cc.CommonCtx, tmpl *template, topazDir string, cfg *clients.Config) *tmplInstaller {
	return &tmplInstaller{
		c:        c,
		tmpl:     tmpl,
		topazDir: topazDir,
		cfg:      cfg,
	}
}

type tmplInstaller struct {
	c        *cc.CommonCtx
	tmpl     *template
	topazDir string
	cfg      *clients.Config
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
	command := DeleteManifestCmd{
		Force:  true,
		Config: *i.cfg,
	}
	return command.Run(i.c)
}

func (i *tmplInstaller) setManifest() error {
	manifest := i.tmpl.AbsURL(i.tmpl.Assets.Manifest)

	if exists, _ := config.FileExists(manifest); !exists {
		manifestDir := path.Join(i.topazDir, "model")
		switch m, err := download(manifest, manifestDir); {
		case err != nil:
			return err
		default:
			manifest = m
		}
	}

	command := SetManifestCmd{
		Path:   manifest,
		Config: *i.cfg,
	}

	return command.Run(i.c)
}

func (i *tmplInstaller) importData() error {
	defaultDataDir := path.Join(i.topazDir, "data")

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
		command := ImportCmd{
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
	assertionsDir := path.Join(i.topazDir, "assertions")

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

	for _, v := range tests {
		command := TestExecCmd{
			File:    v,
			NoColor: false,
			Summary: true,
			Config:  *i.cfg,
		}

		if err := command.Run(i.c); err != nil {
			return err
		}
	}
	return nil
}

func download(src, dir string) (string, error) {
	buf, err := getBytes(src)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}

	f, err := os.Create(filepath.Join(dir, filepath.Base(src)))
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := f.Write(buf); err != nil {
		return "", err
	}

	return f.Name(), nil
}

func getBytes(fileURL string) ([]byte, error) {
	if exists, _ := config.FileExists(fileURL); exists {
		return os.ReadFile(fileURL)
	}

	resp, err := http.Get(fileURL) //nolint used to download the template files
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := bytes.Buffer{}
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func getTemplate(name, templatesURL string) (*template, error) {
	if exists, _ := config.FileExists(name); exists {
		return getTemplateFromFile(name)
	}

	tmplURL, err := url.Parse(templatesURL)
	if err != nil {
		return nil, err
	}

	ref, err := getTemplateRef(name, tmplURL)
	if err != nil {
		return nil, err
	}

	tmpl, err := ref.getTemplate()
	if err != nil {
		return nil, err
	}
	tmpl.Name = name

	return tmpl, nil
}

func getCatalog(templatesURL string) (tmplCatalog, error) {
	buf, err := getBytes(templatesURL)
	if err != nil {
		return nil, err
	}

	var ctlg tmplCatalog
	if err := json.Unmarshal(buf, &ctlg); err != nil {
		return nil, err
	}

	return ctlg, nil
}

func getTemplateRef(name string, templatesURL *url.URL) (*tmplRef, error) {
	list, err := getCatalog(templatesURL.String())
	if err != nil {
		return nil, err
	}

	if ref, ok := list[name]; ok {
		parsed, err := url.Parse(ref.URL)
		if err != nil {
			return nil, err
		}
		if parsed.Scheme == "" {
			absURL := *templatesURL
			absURL.Path = path.Join(path.Dir(absURL.Path), parsed.Path)
			ref.absURL = &absURL
		}

		return ref, nil
	}

	return nil, derr.ErrNotFound.Msgf("template %s", name)
}

func getTemplateFromFile(file string) (*template, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	tmpl := template{base: &url.URL{Path: path.Dir(file)}}
	if err := json.Unmarshal(b, &tmpl); err != nil {
		return nil, err
	}

	return &tmpl, nil
}

func (i *tmplRef) getTemplate() (*template, error) {
	buf, err := getBytes(i.absURL.String())
	if err != nil {
		return nil, err
	}

	base := *i.absURL
	base.Path = path.Dir(base.Path)
	tmpl := template{base: &base}
	if err := json.Unmarshal(buf, &tmpl); err != nil {
		return nil, err
	}

	// TODO: remove after updating all templates to URLs relative to the template definition.
	tmpl.Assets.Manifest = strings.TrimPrefix(tmpl.Assets.Manifest, base.Path)
	for i, v := range tmpl.Assets.IdentityData {
		tmpl.Assets.IdentityData[i] = strings.TrimPrefix(v, base.Path)
	}
	for i, v := range tmpl.Assets.DomainData {
		tmpl.Assets.DomainData[i] = strings.TrimPrefix(v, base.Path)
	}

	return &tmpl, nil
}

func max(rhs, lhs int) int {
	if rhs < lhs {
		return lhs
	}
	return rhs
}

func getTopazDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Clean(path.Join(homeDir, ".config", "topaz")), nil
}

type VerifyTemplateCmd struct {
	TemplatesURL string `arg:"" required:"false" default:"https://topaz.sh/assets/templates/templates.json" help:"URL of template catalog"`
}

func (cmd *VerifyTemplateCmd) Run(c *cc.CommonCtx) error {
	ctlg, err := getCatalog(cmd.TemplatesURL)
	if err != nil {
		return err
	}

	// limit the amount of noise from the azm parser.
	zerolog.SetGlobalLevel(zerolog.Disabled)

	table := c.UI.Normal().WithTable("template", "asset", "exists", "parsed", "error")
	table.WithTableNoAutoWrapText()

	for tmplName := range ctlg {

		tmpl, err := getTemplate(tmplName, cmd.TemplatesURL)
		if err != nil {
			return err
		}
		{
			absURL := tmpl.AbsURL(tmpl.Assets.Manifest)
			exists, parsed, err := validateManifest(absURL)
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			table.WithTableRow(tmplName, absURL, fmt.Sprintf("%t", exists), fmt.Sprintf("%t", parsed), errStr)
		}
		{
			assets := []string{}
			assets = append(assets, tmpl.Assets.Assertions...)
			assets = append(assets, tmpl.Assets.IdentityData...)
			assets = append(assets, tmpl.Assets.DomainData...)

			for _, assetURL := range assets {
				absURL := tmpl.AbsURL(assetURL)
				exists, parsed, err := validateJSON(absURL)
				errStr := ""
				if err != nil {
					errStr = err.Error()
				}
				table.WithTableRow(tmplName, absURL, fmt.Sprintf("%t", exists), fmt.Sprintf("%t", parsed), errStr)
			}
		}
	}

	table.Do()

	return nil
}

func validateManifest(absURL string) (exists, parsed bool, err error) {
	b, err := getBytes(absURL)
	if err != nil {
		return false, false, err
	}

	if _, err := v3.Load(bytes.NewReader(b)); err != nil {
		return true, false, err
	}

	return true, true, nil
}

func validateJSON(absURL string) (exists, parsed bool, err error) {
	b, err := getBytes(absURL)
	if err != nil {
		return false, false, err
	}

	var v map[string]interface{}
	if err := json.NewDecoder(bytes.NewReader(b)).Decode(&v); err != nil {
		return true, false, err
	}

	return true, true, nil
}
