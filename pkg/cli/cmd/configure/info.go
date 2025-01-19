package configure

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/internal/pkg/xdg"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/common"

	"github.com/itchyny/gojq"
)

type InfoConfigCmd struct {
	Var string `arg:"" optional:"" help:"configuration variable"`
	Raw bool   `flag:"" short:"r" help:"output raw strings"`
}

func (cmd InfoConfigCmd) Run(c *cc.CommonCtx) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)

	// use Info struct when output all, to preserve ordering of root objects.
	if cmd.Var == "" {
		return enc.Encode(cmd.info(c))
	}

	query, err := gojq.Parse("." + cmd.Var)
	if err != nil {
		return err
	}

	iter := query.Run(cmd.json(c))
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			if err, ok := err.(*gojq.HaltError); ok && err.Value() == nil {
				break
			}
			return err
		}
		if s, ok := v.(string); ok && cmd.Raw {
			fmt.Fprintln(c.StdOut(), s)
		} else {
			_ = enc.Encode(v)
		}
	}

	return nil
}

type Info struct {
	Environment struct {
		Home          string `json:"home"`
		XdgConfigHome string `json:"xdg_config_home"`
		XdgDataHome   string `json:"xdg_data_home"`
	} `json:"environment"`
	Config struct {
		TopazCfgDir      string `json:"topaz_cfg_dir"`
		TopazCertsDir    string `json:"topaz_certs_dir"`
		TopazDataDir     string `json:"topaz_db_dir"`
		TopazTemplateDir string `json:"topaz_tmpl_dir"`
		TopazDir         string `json:"topaz_dir"`
	} `json:"config"`
	Runtime struct {
		ActiveConfigurationName  string `json:"active_configuration_name"`
		ActiveConfigurationFile  string `json:"active_configuration_file"`
		RunningConfigurationName string `json:"running_configuration_name"`
		RunningConfigurationFile string `json:"running_configuration_file"`
		RunningContainerName     string `json:"running_container_name"`
		TopazConfigFile          string `json:"topaz_json"`
	} `json:"runtime"`
	Default struct {
		ContainerRegistry string `json:"container_registry"`
		ContainerImage    string `json:"container_image"`
		ContainerTag      string `json:"container_tag"`
		ContainerPlatform string `json:"container_platform"`
		NoCheck           bool   `json:"topaz_no_check"`
		NoColor           bool   `json:"topaz_no_color"`
	} `json:"default"`
	Directory struct {
		DirectorySvc   string `json:"topaz_directory_svc"`
		DirectoryKey   string `json:"topaz_directory_key"`
		DirectoryToken string `json:"topaz_directory_token"`
		Insecure       bool   `json:"topaz_insecure"`
		TenantID       string `json:"aserto_tenant_id"`
	} `json:"directory"`
	Authorizer struct {
		AuthorizerSvc   string `json:"topaz_authorizer_svc"`
		AuthorizerKey   string `json:"topaz_authorizer_key"`
		AuthorizerToken string `json:"topaz_authorizer_token"`
		Insecure        bool   `json:"topaz_insecure"`
		TenantID        string `json:"aserto_tenant_id"`
	} `json:"authorizer"`
}

func (cmd InfoConfigCmd) info(c *cc.CommonCtx) *Info {
	info := Info{}

	info.Environment.Home = xdg.Home
	info.Environment.XdgConfigHome = xdg.ConfigHome
	info.Environment.XdgDataHome = xdg.DataHome

	info.Config.TopazCfgDir = cc.GetTopazCfgDir()
	info.Config.TopazCertsDir = cc.GetTopazCertsDir()
	info.Config.TopazDataDir = cc.GetTopazDataDir()
	info.Config.TopazTemplateDir = cc.GetTopazTemplateDir()
	info.Config.TopazDir = cc.GetTopazDir()

	info.Runtime.ActiveConfigurationName = c.Config.Active.Config
	info.Runtime.ActiveConfigurationFile = c.Config.Active.ConfigFile
	info.Runtime.RunningConfigurationName = c.Config.Running.Config
	info.Runtime.RunningConfigurationFile = c.Config.Running.ConfigFile
	info.Runtime.RunningContainerName = c.Config.Running.ContainerName
	info.Runtime.TopazConfigFile = filepath.Join(cc.GetTopazDir(), common.CLIConfigurationFile)

	info.Default.ContainerRegistry = cc.ContainerRegistry()
	info.Default.ContainerImage = cc.ContainerImage()
	info.Default.ContainerTag = cc.ContainerTag()
	info.Default.ContainerPlatform = cc.ContainerPlatform()
	info.Default.NoCheck = cc.NoCheck()
	info.Default.NoColor = cc.NoColor()

	info.Directory.DirectorySvc = cc.DirectorySvc()
	info.Directory.DirectoryKey = cc.DirectoryKey()
	info.Directory.DirectoryToken = cc.DirectoryToken()
	info.Directory.Insecure = cc.Insecure()
	info.Directory.TenantID = cc.TenantID()

	info.Authorizer.AuthorizerSvc = cc.AuthorizerSvc()
	info.Authorizer.AuthorizerKey = cc.AuthorizerKey()
	info.Authorizer.AuthorizerToken = cc.AuthorizerToken()
	info.Authorizer.Insecure = cc.Insecure()
	info.Authorizer.TenantID = cc.TenantID()

	return &info
}

func (cmd InfoConfigCmd) json(c *cc.CommonCtx) map[string]any {
	var j map[string]any
	buf, _ := json.Marshal(cmd.info(c))
	_ = json.Unmarshal(buf, &j)
	return j
}
