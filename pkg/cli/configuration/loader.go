package configuration

import (
	"bytes"
	"net"
	"os"

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

	for _, service := range l.Configuration.APIConfig.Services {
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
