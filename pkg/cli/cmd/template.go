package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/go-directory/pkg/derr"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
)

const (
	templatesURL   string = "https://topaz.sh/assets/templates.json"
	colName        string = "name"
	colDescription string = "description"
)

type tmplList map[string]*tmplItem

type tmplItem struct {
	Description string `json:"description"`
	URL         string `json:"url"`
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

type ListTemplatesCmd struct{}

func (cmd *ListTemplatesCmd) Run(c *cc.CommonCtx) error {
	buf, err := getBytesFromURL("https://topaz.sh/assets/templates.json")
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

	fmt.Fprintf(c.UI.Output(), "%-*s : %s\n", maxWidth, colName, colDescription)
	fmt.Fprintf(c.UI.Output(), "%-*s : %s\n", maxWidth, strings.Repeat("-", maxWidth), strings.Repeat("-", len(colDescription)))

	for n, t := range tmplList {
		fmt.Fprintf(c.UI.Output(), "%-*s : %s\n", maxWidth, n, t.Description)
	}

	return nil
}

type InstallTemplateCmd struct {
	Name string `arg:"" required:"" help:"template name"`
	clients.Config
}

func (cmd *InstallTemplateCmd) Run(c *cc.CommonCtx) error {
	item, err := getItem(cmd.Name)
	if err != nil {
		return err
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
// 9 - topaz console, launch console so the user start exploring the template artifacts
func (cmd *InstallTemplateCmd) installTemplate(c *cc.CommonCtx, i *tmplInstance) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	topazDir := path.Join(homeDir, ".config", "topaz")

	os.Setenv("TOPAZ_DIR", topazDir)

	cmd.Config.Insecure = true

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
		}
		if err := command.Run(c); err != nil {
			return err
		}
	}
	// 3 - topaz start - start instance using new configuration
	{
		command := &StartCmd{
			ContainerName:    "topaz",
			ContainerVersion: "latest",
		}
		if err := command.Run(c); err != nil {
			return err
		}
	}
	// 4 - wait for health endpoint to be in serving state
	{
		if !isServing() {
			return fmt.Errorf("gRPC endpoint not SERVING")
		}
	}
	// 5 - topaz manifest delete --force, reset the directory store
	{
		command := DeleteManifestCmd{}
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

		buf, err := getBytesFromURL(i.Assets.Manifest)
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
			buf, err := getBytesFromURL(v)
			if err != nil {
				return err
			}

			f, err := os.Create(filepath.Join(dataDir, filepath.Base(v)))
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := f.Write(buf); err != nil {
				return err
			}
		}

		for _, v := range i.Assets.DomainData {
			buf, err := getBytesFromURL(v)
			if err != nil {
				return err
			}

			f, err := os.Create(filepath.Join(dataDir, filepath.Base(v)))
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := f.Write(buf); err != nil {
				return err
			}
		}

		command := ImportCmd{
			Directory: dataDir,
			Config:    cmd.Config,
		}

		if err := command.Run(c); err != nil {
			return err
		}
	}
	// 8 - topaz test exec, execute assertions when part of template
	{
		assertionsDir := path.Join(topazDir, "assertions")
		if err := os.MkdirAll(assertionsDir, 0700); err != nil {
			return err
		}

		for _, v := range i.Assets.Assertions {
			buf, err := getBytesFromURL(v)
			if err != nil {
				return err
			}

			f, err := os.Create(filepath.Join(assertionsDir, filepath.Base(v)))
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := f.Write(buf); err != nil {
				return err
			}
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
	}
	// 9 - topaz console, launch console so the user start exploring the template artifacts
	{
		command := ConsoleCmd{
			ConsoleAddress: "https://localhost:8080/ui/directory",
		}
		if err := command.Run(c); err != nil {
			return err
		}
	}

	return nil
}

func getBytesFromURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
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

func getTemplateList() (tmplList, error) {
	buf, err := getBytesFromURL(templatesURL)
	if err != nil {:
		return nil, err
	}

	var tmplList map[string]*tmplItem
	if err := json.Unmarshal(buf, &tmplList); err != nil {
		return nil, err
	}

	return tmplList, nil
}

func getItem(name string) (*tmplItem, error) {
	list, err := getTemplateList()
	if err != nil {
		return nil, err
	}

	if item, ok := list[name]; ok {
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
