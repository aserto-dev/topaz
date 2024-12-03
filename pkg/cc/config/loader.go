package config

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/aserto-dev/self-decision-logger/logger/self"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	builder "github.com/aserto-dev/topaz/pkg/service/builder"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Loader struct {
	Configuration *Config
	HasTopazDir   bool
}

var envRegex = regexp.MustCompile(`(?U:\${.*})`)

type replacer struct {
	r *strings.Replacer
}

func newReplacer() *replacer {
	return &replacer{r: strings.NewReplacer(".", "_")}
}

func (r replacer) Replace(s string) string {
	if s == "TOPAZ_VERSION" {
		// Prevent the `version` field from be overridden by env vars.
		return ""
	}

	return r.r.Replace(s)
}

func LoadConfiguration(fileName string) (*Loader, error) {
	v := viper.NewWithOptions(viper.EnvKeyReplacer(newReplacer()))
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.SetConfigFile(fileName)
	v.SetEnvPrefix("TOPAZ")

	// Set defaults
	v.SetDefault("debug_service.enabled", false)
	v.SetDefault("debug_service.listen_address", "localhost:6060")
	v.SetDefault("debug_service.shutdown_timeout", 0)

	v.SetDefault("jwt.acceptable_time_skew_seconds", 5)

	v.SetDefault("opa.max_plugin_wait_time_seconds", "30")

	v.SetDefault("remote_directory.address", "0.0.0.0:9292")
	v.SetDefault("remote_directory.insecure", "true")

	v.AutomaticEnv()

	fileContents, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	withTopazDir := false
	if strings.Contains(string(fileContents), "TOPAZ_DIR") {
		withTopazDir = true
	}
	cfg := new(Config)
	subBuf, err := SetEnvVars(string(fileContents))
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader([]byte(subBuf))

	if err := v.ReadConfig(r); err != nil {
		return nil, err
	}
	err = v.UnmarshalExact(cfg, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "json"
	})
	if err != nil {
		return nil, err
	}

	return &Loader{
		Configuration: cfg,
		HasTopazDir:   withTopazDir,
	}, nil
}

func (l *Loader) GetPaths() ([]string, error) {
	paths := make(map[string]bool)

	if l.Configuration.Edge.DBPath != "" {
		paths[l.Configuration.Edge.DBPath] = true
	}
	if l.Configuration.APIConfig.Health.Certificates != nil {
		if l.Configuration.APIConfig.Health.Certificates.CA != "" {
			paths[l.Configuration.APIConfig.Health.Certificates.CA] = true
		}
		if l.Configuration.APIConfig.Health.Certificates.Cert != "" {
			paths[l.Configuration.APIConfig.Health.Certificates.Cert] = true
		}
		if l.Configuration.APIConfig.Health.Certificates.Key != "" {
			paths[l.Configuration.APIConfig.Health.Certificates.Key] = true
		}
	}

	if l.Configuration.APIConfig.Metrics.Certificates != nil {
		if l.Configuration.APIConfig.Metrics.Certificates.CA != "" {
			paths[l.Configuration.APIConfig.Metrics.Certificates.CA] = true
		}
		if l.Configuration.APIConfig.Metrics.Certificates.Cert != "" {
			paths[l.Configuration.APIConfig.Metrics.Certificates.Cert] = true
		}
		if l.Configuration.APIConfig.Metrics.Certificates.Key != "" {
			paths[l.Configuration.APIConfig.Metrics.Certificates.Key] = true
		}
	}

	servicePaths := getUniqueServiceCertPaths(l.Configuration.APIConfig.Services)
	for i := range servicePaths {
		paths[servicePaths[i]] = true
	}

	if l.Configuration.ControllerConfig != nil && l.Configuration.ControllerConfig.Enabled {
		if l.Configuration.ControllerConfig.Server.CACertPath != "" {
			paths[l.Configuration.ControllerConfig.Server.CACertPath] = true
		}
		if l.Configuration.ControllerConfig.Server.ClientCertPath != "" {
			paths[l.Configuration.ControllerConfig.Server.ClientCertPath] = true
		}
		if l.Configuration.ControllerConfig.Server.ClientKeyPath != "" {
			paths[l.Configuration.ControllerConfig.Server.ClientKeyPath] = true
		}
	}

	decisionLogPaths, err := getDecisionLogPaths(l.Configuration.DecisionLogger)
	if err != nil {
		return nil, err
	}
	for i := range decisionLogPaths {
		paths[decisionLogPaths[i]] = true
	}

	return filterPaths(paths), nil
}

func (l *Loader) GetPorts() ([]string, error) {
	portMap := make(map[string]bool)

	if l.Configuration.APIConfig.Health.ListenAddress != "" {
		port, err := getPortFromAddress(l.Configuration.APIConfig.Health.ListenAddress)
		if err != nil {
			return nil, err
		}
		portMap[port] = true
	}

	if l.Configuration.APIConfig.Metrics.ListenAddress != "" {
		port, err := getPortFromAddress(l.Configuration.APIConfig.Metrics.ListenAddress)
		if err != nil {
			return nil, err
		}
		portMap[port] = true
	}

	if l.Configuration.DebugService.Enabled && l.Configuration.DebugService.ListenAddress != "" {
		port, err := getPortFromAddress(l.Configuration.DebugService.ListenAddress)
		if err != nil {
			return nil, err
		}
		portMap[port] = true
	}

	for _, value := range l.Configuration.APIConfig.Services {
		if value.GRPC.ListenAddress != "" {
			port, err := getPortFromAddress(value.GRPC.ListenAddress)
			if err != nil {
				return nil, err
			}
			portMap[port] = true
		}

		if value.Gateway.ListenAddress != "" {
			port, err := getPortFromAddress(value.Gateway.ListenAddress)
			if err != nil {
				return nil, err
			}
			portMap[port] = true
		}
	}

	// ensure unique assignment for each port
	var args []string
	for k := range portMap {
		args = append(args, k)
	}
	return args, nil
}

func SetEnvVars(fileContents string) (string, error) {
	err := os.Setenv("TOPAZ_CFG_DIR", cc.GetTopazCfgDir())
	if err != nil {
		return "", err
	}
	err = os.Setenv("TOPAZ_CERTS_DIR", cc.GetTopazCertsDir())
	if err != nil {
		return "", err
	}
	err = os.Setenv("TOPAZ_DB_DIR", cc.GetTopazDataDir())
	if err != nil {
		return "", err
	}
	return subEnvVars(fileContents), nil
}

func filterPaths(paths map[string]bool) []string {
	var args []string
	for k := range paths {
		if k != "" { // append only not empty paths.
			args = append(args, k)
		}
	}
	return args
}

func getPortFromAddress(address string) (string, error) {
	_, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", err
	}
	return port, nil
}

func getUniqueServiceCertPaths(services map[string]*builder.API) []string {
	paths := make(map[string]bool)
	for _, service := range services {
		if service.GRPC.Certs.CA != "" {
			paths[service.GRPC.Certs.CA] = true
		}
		if service.GRPC.Certs.Cert != "" {
			paths[service.GRPC.Certs.Cert] = true
		}
		if service.GRPC.Certs.Key != "" {
			paths[service.GRPC.Certs.Key] = true
		}
		if service.Gateway.Certs.CA != "" {
			paths[service.Gateway.Certs.CA] = true
		}
		if service.Gateway.Certs.Cert != "" {
			paths[service.Gateway.Certs.Cert] = true
		}
		if service.Gateway.Certs.Key != "" {
			paths[service.Gateway.Certs.Key] = true
		}
	}
	var pathList []string
	for k := range paths {
		pathList = append(pathList, k)
	}
	return pathList
}

func getDecisionLogPaths(decisionLogConfig DecisionLogConfig) ([]string, error) {
	switch decisionLogConfig.Type {
	case "file":
		logpath := fmt.Sprintf("%s", decisionLogConfig.Config["log_file_path"])
		return []string{logpath}, nil
	case "self":
		selfCfg := self.Config{}
		dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Result:  &selfCfg,
			TagName: "json",
		})
		if err != nil {
			return nil, err
		}
		err = dec.Decode(decisionLogConfig.Config)
		if err != nil {
			return nil, err
		}
		return []string{selfCfg.StoreDirectory, selfCfg.Scribe.CACertPath, selfCfg.Scribe.ClientCertPath, selfCfg.Scribe.ClientKeyPath}, nil
	default:
		return nil, nil // nop decision logger
	}
}

// subEnvVars will look for any environment variables in the passed in string
// with the syntax of ${VAR_NAME} and replace that string with ENV[VAR_NAME].
func subEnvVars(s string) string {
	updatedConfig := envRegex.ReplaceAllStringFunc(strings.ReplaceAll(s, `"`, `'`), func(s string) string {
		// Trim off the '${' and '}'
		if len(s) <= 3 {
			// This should never happen..
			return ""
		}
		varName := s[2 : len(s)-1]

		// Lookup the variable in the environment. We play by
		// bash rules.. if its undefined we'll treat it as an
		// empty string instead of raising an error.
		return os.Getenv(varName)
	})

	return updatedConfig
}
