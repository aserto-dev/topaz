package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/aserto-dev/go-directory/pkg/derr"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
)

const (
	colName          string = "name"
	colDescription   string = "description"
	colDocumentation string = "documentation"
)

type tmplList map[string]*tmplItem

type tmplItem struct {
	Title            string `json:"title,omitempty"`
	ShortDescription string `json:"short_description,omitempty"`
	Description      string `json:"description,omitempty"`
	URL              string `json:"topaz_url,omitempty"`
	DocumentationURL string `json:"docs_url,omitempty"`
}

type tmplInstance struct {
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
}

type TemplateCmd struct {
	List    ListTemplatesCmd   `cmd:"" help:"list template"`
	Install InstallTemplateCmd `cmd:"" help:"install template"`
}

type ListTemplatesCmd struct {
	TemplatesURL string `arg:"" required:"false" default:"https://topaz.sh" help:"template url"`
}

func (cmd *ListTemplatesCmd) Run(c *cc.CommonCtx) error {
	buf, err := getBytesFromURL(fmt.Sprintf("%s/assets/templates/templates.json", cmd.TemplatesURL))
	if err != nil {
		return err
	}

	var tmplList map[string]*tmplItem

	if err := json.Unmarshal(buf, &tmplList); err != nil {
		return err
	}

	maxWidth := 0
	for n := range tmplList {
		maxWidth = max(maxWidth, len(n)+1)
	}

	table := c.UI.Normal().WithTable(colName, colDescription, colDocumentation)
	table.WithTableNoAutoWrapText()
	for n, t := range tmplList {
		table.WithTableRow(n, t.ShortDescription, t.DocumentationURL)
	}
	table.Do()

	return nil
}

type InstallTemplateCmd struct {
	Name             string `arg:"" required:"" help:"template name"`
	Force            bool   `flag:"" default:"false" required:"false" help:"forcefully apply template"`
	ContainerName    string `optional:"" default:"topaz" help:"container name"`
	ContainerVersion string `optional:"" default:"latest" help:"container version" `
	TemplatesURL     string `arg:"" required:"false" default:"https://topaz.sh" help:"template url"`

	clients.Config
}

func (cmd *InstallTemplateCmd) Run(c *cc.CommonCtx) error {
	item, err := getItem(cmd.Name, fmt.Sprintf("%s/assets/templates/templates.json", cmd.TemplatesURL))
	if err != nil {
		return err
	}
	if !cmd.Force {
		c.UI.Exclamation().Msg("Installing this template will completely reset your topaz configuration.")
		if !promptYesNo("Do you want to continue?", false) {
			return nil
		}
	}

	instance, err := item.getInstance()
	if err != nil {
		return err
	}
	instance.Name = cmd.Name

	return cmd.installTemplate(c, instance)
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
func (cmd *InstallTemplateCmd) installTemplate(c *cc.CommonCtx, i *tmplInstance) error {
	topazDir := GetTopazDir()
	os.Setenv("TOPAZ_DIR", topazDir)

	cmd.Config.Insecure = true
	// prepare Topaz configuration and start container
	if err := cmd.prepareTopaz(c, i); err != nil {
		return err
	}

	// 4 - wait for health endpoint to be in serving state
	if !isServing() {
		return fmt.Errorf("gRPC endpoint not SERVING")
	}

	// reset directory store and load data from template
	if err := cmd.loadData(c, i, topazDir); err != nil {
		return err
	}

	// 8 - topaz test exec, execute assertions when part of template
	if err := cmd.runTemplateTests(c, i, topazDir); err != nil {
		return err
	}

	// 9 - topaz console, launch console so the user start exploring the template artifacts
	command := ConsoleCmd{
		ConsoleAddress: "https://localhost:8080/ui/directory",
	}
	if err := command.Run(c); err != nil {
		return err
	}

	return nil
}

func (cmd *InstallTemplateCmd) prepareTopaz(c *cc.CommonCtx, i *tmplInstance) error {

	// 1 - topaz stop - ensure topaz is not running, so we can reconfigure
	{
		command := &StopCmd{}
		if err := command.Run(c); err != nil {
			return err
		}
	}

	// 2 - topaz configure - generate a new configuration based on the requirements of the template
	{
		command := ConfigureCmd{
			PolicyName: i.Assets.Policy.Name,
			Resource:   i.Assets.Policy.Resource,
			Force:      true,
		}
		if err := command.Run(c); err != nil {
			return err
		}
	}
	// 3 - topaz start - start instance using new configuration
	{
		command := &StartCmd{
			ContainerName:    cmd.ContainerName,
			ContainerVersion: cmd.ContainerVersion,
		}
		if err := command.Run(c); err != nil {
			return err
		}
	}
	return nil
}

func (cmd *InstallTemplateCmd) loadData(c *cc.CommonCtx, i *tmplInstance, topazDir string) error {
	// 5 - topaz manifest delete --force, reset the directory store
	{
		command := DeleteManifestCmd{
			Force:  true,
			Config: cmd.Config,
		}
		if err := command.Run(c); err != nil {
			return err
		}
	}
	// 6 - topaz manifest set, deploy the manifest
	{
		manifestDir := path.Join(topazDir, "model")
		if err := os.MkdirAll(manifestDir, 0700); err != nil {
			return err
		}

		buf, err := getBytesFromURL(fmt.Sprintf("%s/%s", cmd.TemplatesURL, i.Assets.Manifest))
		if err != nil {
			return err
		}

		manifest := filepath.Join(manifestDir, filepath.Base(i.Assets.Manifest))
		f, err := os.Create(manifest)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := f.Write(buf); err != nil {
			return err
		}

		command := SetManifestCmd{
			Path:   manifest,
			Stdin:  false,
			Config: cmd.Config,
		}

		if err := command.Run(c); err != nil {
			return err
		}
	}
	// 7 - topaz import, load IDP and domain data (in that order)
	{
		dataDir := path.Join(topazDir, "data")
		if err := os.MkdirAll(dataDir, 0700); err != nil {
			return err
		}

		for _, v := range i.Assets.IdentityData {
			buf, err := getBytesFromURL(fmt.Sprintf("%s/%s", cmd.TemplatesURL, v))
			if err != nil {
				return err
			}

			f, err := os.Create(filepath.Join(dataDir, filepath.Base(v)))
			if err != nil {
				return err
			}

			if _, err := f.Write(buf); err != nil {
				return err
			}
			f.Close()
		}

		for _, v := range i.Assets.DomainData {
			buf, err := getBytesFromURL(fmt.Sprintf("%s/%s", cmd.TemplatesURL, v))
			if err != nil {
				return err
			}

			f, err := os.Create(filepath.Join(dataDir, filepath.Base(v)))
			if err != nil {
				return err
			}

			if _, err := f.Write(buf); err != nil {
				return err
			}
			f.Close()
		}

		command := ImportCmd{
			Directory: dataDir,
			Config:    cmd.Config,
		}

		if err := command.Run(c); err != nil {
			return err
		}
	}
	return nil
}

func (cmd *InstallTemplateCmd) runTemplateTests(c *cc.CommonCtx, i *tmplInstance, topazDir string) error {
	assertionsDir := path.Join(topazDir, "assertions")
	if err := os.MkdirAll(assertionsDir, 0700); err != nil {
		return err
	}

	for _, v := range i.Assets.Assertions {
		buf, err := getBytesFromURL(fmt.Sprintf("%s/%s", cmd.TemplatesURL, v))
		if err != nil {
			return err
		}

		f, err := os.Create(filepath.Join(assertionsDir, filepath.Base(v)))
		if err != nil {
			return err
		}

		if _, err := f.Write(buf); err != nil {
			return err
		}
		f.Close()
	}

	for _, v := range i.Assets.Assertions {
		command := TestExecCmd{
			File:    filepath.Join(assertionsDir, filepath.Base(v)),
			NoColor: false,
			Summary: true,
			Config:  cmd.Config,
		}

		if err := command.Run(c); err != nil {
			return err
		}
	}
	return nil
}

func getBytesFromURL(fileURL string) ([]byte, error) {
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

func getTemplateList(templatesURL string) (tmplList, error) {
	buf, err := getBytesFromURL(templatesURL)
	if err != nil {
		return nil, err
	}

	var tmplList map[string]*tmplItem
	if err := json.Unmarshal(buf, &tmplList); err != nil {
		return nil, err
	}

	return tmplList, nil
}

func getItem(name, templatesURL string) (*tmplItem, error) {
	list, err := getTemplateList(templatesURL)
	if err != nil {
		return nil, err
	}

	if item, ok := list[name]; ok {
		parsed, err := url.Parse(item.URL)
		if err != nil {
			return nil, err
		}
		if parsed.Scheme == "" {
			base := strings.Replace(templatesURL, path.Base(templatesURL), "", 1)
			fullURL := base + "/" + item.URL
			item.URL = fullURL
		}

		return item, nil
	}

	return nil, derr.ErrNotFound.Msgf("template %s", name)
}

func (i *tmplItem) getInstance() (*tmplInstance, error) {
	buf, err := getBytesFromURL(i.URL)
	if err != nil {
		return nil, err
	}

	var instance tmplInstance
	if err := json.Unmarshal(buf, &instance); err != nil {
		return nil, err
	}

	return &instance, nil
}

func max(rhs, lhs int) int {
	if rhs < lhs {
		return lhs
	}
	return rhs
}

func isServing() bool {
	cmd := exec.Command("grpc-health-probe", "-addr=localhost:9494", "-connect-timeout=30s", "-rpc-timeout=30s")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	fmt.Println(string(out))
	return true
}

func GetTopazDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Clean(filepath.Join(getHomeDefault(), ".config", "topaz"))
	}
	return filepath.Clean(filepath.Join(homeDir, ".config", "topaz"))
}

func GetTopazCfgDir() string {
	return filepath.Clean(filepath.Join(GetTopazDir(), "cfg"))
}

func GetTopazCertsDir() string {
	return filepath.Clean(filepath.Join(GetTopazDir(), "certs"))
}

const (
	fallback        = `~/.config/topaz`
	darwinFallBack  = fallback
	linuxFallBack   = fallback
	windowsFallBack = `\\.config\\topaz`
)

func getHomeDefault() string {
	switch runtime.GOOS {
	case "darwin":
		return darwinFallBack
	case "linux":
		return linuxFallBack
	case "windows":
		return filepath.Join(os.Getenv("USERPROFILE"), windowsFallBack)
	default:
		return ""
	}
}
