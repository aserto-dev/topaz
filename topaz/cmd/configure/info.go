package configure

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/topaz/internal/xdg"
	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/topaz/cc"
	"github.com/aserto-dev/topaz/topaz/cmd/common"
	"github.com/aserto-dev/topaz/topazd/service/builder"
	"github.com/samber/lo"

	"github.com/itchyny/gojq"
)

type InfoConfigCmd struct {
	Var string `arg:"" optional:"" help:"configuration variable"`
	Raw bool   `flag:"" short:"r" help:"output raw strings"`
}

func (cmd InfoConfigCmd) Run(ctx context.Context) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)

	// use Info struct when output all, to preserve ordering of root objects.
	if cmd.Var == "" {
		return enc.Encode(cmd.GetInfo())
	}

	query, err := gojq.Parse("." + cmd.Var)
	if err != nil {
		return err
	}

	iter := query.Run(cmd.json())

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}

		if err, ok := v.(error); ok {
			if err, ok := err.(*gojq.HaltError); ok && err.Value() == nil { //nolint:errorlint
				break
			}

			return err
		}

		if s, ok := v.(string); ok && cmd.Raw {
			fmt.Fprintln(os.Stdout, s)
		} else {
			_ = enc.Encode(v) //nolint:errchkjson // ok
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
		TopazCfgDir       string `json:"topaz_cfg_dir"`
		TopazCertsDir     string `json:"topaz_certs_dir"`
		TopazDataDir      string `json:"topaz_db_dir"`
		TopazDecisionsDir string `json:"topaz_decisions_dir"`
		TopazTemplateDir  string `json:"topaz_tmpl_dir"`
		TopazDir          string `json:"topaz_dir"`
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
		DirectorySvc     string `json:"topaz_directory_svc"`
		DirectorySvcHttp string `json:"topaz_directory_svc_http"`
		DirectoryKey     string `json:"topaz_directory_key"`
		DirectoryToken   string `json:"topaz_directory_token"`
		Insecure         bool   `json:"topaz_insecure"`
		Plaintext        bool   `json:"topaz_plaintext"`
		Timeout          string `json:"timeout"`
	} `json:"directory"`
	Authorizer struct {
		AuthorizerSvc     string `json:"topaz_authorizer_svc"`
		AuthorizerSvcHttp string `json:"topaz_authorizer_svc_http"`
		AuthorizerKey     string `json:"topaz_authorizer_key"`
		AuthorizerToken   string `json:"topaz_authorizer_token"`
		Insecure          bool   `json:"topaz_insecure"`
		Plaintext         bool   `json:"topaz_plaintext"`
		Timeout           string `json:"timeout"`
	} `json:"authorizer"`
	OpenAPI struct {
		Directory  string `json:"topaz_directory_svc"`
		Access     string `json:"topaz_access_svc"`
		Authorizer string `json:"topaz_authorizer_svc"`
	} `json:"open_api_endpoints"`
	Access struct {
		WellKnownConfigURL string `json:"wellknown_config_url"`
	} `json:"access"`
}

func (cmd InfoConfigCmd) GetInfo() *Info {
	info := Info{}

	info.Environment.Home = xdg.Home
	info.Environment.XdgConfigHome = xdg.ConfigHome
	info.Environment.XdgDataHome = xdg.DataHome

	info.Config.TopazCfgDir = cc.GetTopazCfgDir()
	info.Config.TopazCertsDir = cc.GetTopazCertsDir()
	info.Config.TopazDataDir = cc.GetTopazDataDir()
	info.Config.TopazDecisionsDir = cc.GetTopazDecisionsDir()
	info.Config.TopazTemplateDir = cc.GetTopazTemplateDir()
	info.Config.TopazDir = cc.GetTopazDir()

	cfg := cc.GetConfig()
	info.Runtime.ActiveConfigurationName = cfg.Active.Config
	info.Runtime.ActiveConfigurationFile = cfg.Active.ConfigFile
	info.Runtime.RunningConfigurationName = cfg.Running.Config
	info.Runtime.RunningConfigurationFile = cfg.Running.ConfigFile
	info.Runtime.RunningContainerName = cfg.Running.ContainerName
	info.Runtime.TopazConfigFile = filepath.Join(cc.GetTopazDir(), common.CLIConfigurationFile)

	svcCfg := cmd.svcConfig()

	config.GetConfig(cfg.Running.ConfigFile)

	info.Default.ContainerRegistry = cc.ContainerRegistry()
	info.Default.ContainerImage = cc.ContainerImage()
	info.Default.ContainerTag = cc.ContainerTag()
	info.Default.ContainerPlatform = cc.ContainerPlatform()
	info.Default.NoCheck = cc.NoCheck()
	info.Default.NoColor = cc.NoColor()

	info.Directory.DirectorySvc = grpcEndpoint(svcCfg.APIConfig.Services["reader"])
	info.Directory.DirectorySvcHttp = httpEndpoint(svcCfg.APIConfig.Services["reader"])
	info.Directory.DirectoryKey = cc.DirectoryKey()
	info.Directory.DirectoryToken = cc.DirectoryToken()
	info.Directory.Insecure = cc.Insecure()
	info.Directory.Plaintext = cc.Plaintext()
	info.Directory.Timeout = cc.Timeout().String()

	info.Authorizer.AuthorizerSvc = grpcEndpoint(svcCfg.APIConfig.Services["authorizer"])
	info.Authorizer.AuthorizerSvcHttp = httpEndpoint(svcCfg.APIConfig.Services["authorizer"])
	info.Authorizer.AuthorizerKey = cc.AuthorizerKey()
	info.Authorizer.AuthorizerToken = cc.AuthorizerToken()
	info.Authorizer.Insecure = cc.Insecure()
	info.Authorizer.Plaintext = cc.Plaintext()
	info.Authorizer.Timeout = cc.Timeout().String()

	info.OpenAPI.Directory = info.Directory.DirectorySvcHttp + "/directory/openapi.json"
	info.OpenAPI.Access = info.Directory.DirectorySvcHttp + "/access/openapi.json"
	info.OpenAPI.Authorizer = info.Authorizer.AuthorizerSvcHttp + "/authorizer/openapi.json"

	info.Access.WellKnownConfigURL = info.Directory.DirectorySvcHttp + "/.well-known/authzen-configuration"

	return &info
}

func (cmd InfoConfigCmd) json() map[string]any {
	var j map[string]any

	buf, err := json.Marshal(cmd.GetInfo())
	if err != nil {
		return map[string]any{}
	}

	if err := json.Unmarshal(buf, &j); err != nil {
		return map[string]any{}
	}

	return j
}

func (cmd InfoConfigCmd) svcConfig() *config.Config {
	cfg := cc.GetConfig()

	if cfg.Running.ConfigFile != "" {
		c := config.GetConfig(cfg.Running.ConfigFile)
		return c.Configuration
	}

	if cfg.Active.ConfigFile != "" {
		c := config.GetConfig(cfg.Active.ConfigFile)
		return c.Configuration
	}

	return nil
}

func grpcEndpoint(cfg *builder.API) string {
	if cfg == nil {
		return ""
	}

	return lo.Ternary(cfg.GRPC.FQDN != "", cfg.GRPC.FQDN, cfg.GRPC.ListenAddress)
}

func httpEndpoint(cfg *builder.API) string {
	if cfg == nil {
		return ""
	}

	strURL := fmt.Sprintf(
		"%s://%s",
		lo.Ternary(cfg.Gateway.HTTP, "http", "https"),
		strings.Replace(
			lo.Ternary(cfg.Gateway.FQDN != "", cfg.Gateway.FQDN, cfg.Gateway.ListenAddress),
			"0.0.0.0", "localhost", 1,
		),
	)

	u, err := url.Parse(strURL)
	if err != nil {
		return ""
	}

	return u.String()
}
