package configuration

import (
	"bytes"
	"fmt"
	"net"
	"os"

	"github.com/aserto-dev/self-decision-logger/logger/self"
	builder "github.com/aserto-dev/service-host"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Loader struct {
	Configuration *config.Config
}

func LoadConfiguration(fileName string) (*Loader, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.SetConfigFile(fileName)
	fileContents, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	cfg := new(config.Config)

	r := bytes.NewReader(fileContents)
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
	}, nil
}

func (l *Loader) GetPaths() ([]string, error) {
	paths := make(map[string]bool)

	if l.Configuration.Edge.DBPath != "" {
		paths[l.Configuration.Edge.DBPath] = true
	}
	if l.Configuration.APIConfig.Health.Certificates != nil {
		if l.Configuration.APIConfig.Health.Certificates.TLSCACertPath != "" {
			paths[l.Configuration.APIConfig.Health.Certificates.TLSCACertPath] = true
		}
		if l.Configuration.APIConfig.Health.Certificates.TLSCertPath != "" {
			paths[l.Configuration.APIConfig.Health.Certificates.TLSCertPath] = true
		}
		if l.Configuration.APIConfig.Health.Certificates.TLSKeyPath != "" {
			paths[l.Configuration.APIConfig.Health.Certificates.TLSKeyPath] = true
		}
	}

	if l.Configuration.APIConfig.Metrics.Certificates != nil {
		if l.Configuration.APIConfig.Metrics.Certificates.TLSCACertPath != "" {
			paths[l.Configuration.APIConfig.Metrics.Certificates.TLSCACertPath] = true
		}
		if l.Configuration.APIConfig.Metrics.Certificates.TLSCertPath != "" {
			paths[l.Configuration.APIConfig.Metrics.Certificates.TLSCertPath] = true
		}
		if l.Configuration.APIConfig.Metrics.Certificates.TLSKeyPath != "" {
			paths[l.Configuration.APIConfig.Metrics.Certificates.TLSKeyPath] = true
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

	var args []string
	for k := range paths {
		args = append(args, k)
	}
	return args, nil
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
		if service.GRPC.Certs.TLSCACertPath != "" {
			paths[service.GRPC.Certs.TLSCACertPath] = true
		}
		if service.GRPC.Certs.TLSCertPath != "" {
			paths[service.GRPC.Certs.TLSCertPath] = true
		}
		if service.GRPC.Certs.TLSKeyPath != "" {
			paths[service.GRPC.Certs.TLSKeyPath] = true
		}
		if service.Gateway.Certs.TLSCACertPath != "" {
			paths[service.Gateway.Certs.TLSCACertPath] = true
		}
		if service.Gateway.Certs.TLSCertPath != "" {
			paths[service.Gateway.Certs.TLSCertPath] = true
		}
		if service.Gateway.Certs.TLSKeyPath != "" {
			paths[service.Gateway.Certs.TLSKeyPath] = true
		}
	}
	var pathList []string
	for k := range paths {
		pathList = append(pathList, k)
	}
	return pathList
}

func getDecisionLogPaths(decisionLogConfig config.DecisionLogConfig) ([]string, error) {
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
		err = dec.Decode(decisionLogConfig)
		if err != nil {
			return nil, err
		}
		return []string{selfCfg.StoreDirectory, selfCfg.Scribe.CACertPath, selfCfg.Scribe.ClientCertPath, selfCfg.Scribe.ClientKeyPath}, nil
	default:
		return nil, nil // nop decision logger

	}
}
