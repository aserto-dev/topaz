package config_test

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/aserto-dev/aserto-management/controller"
	"github.com/aserto-dev/topaz/pkg/authentication"
	"github.com/aserto-dev/topaz/pkg/authorizer"
	config2 "github.com/aserto-dev/topaz/pkg/cc/config"
	config3 "github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/pkg/debug"
	"github.com/aserto-dev/topaz/pkg/directory"
	"github.com/aserto-dev/topaz/pkg/health"
	"github.com/aserto-dev/topaz/pkg/metrics"
	"github.com/aserto-dev/topaz/pkg/service/builder"
	"github.com/aserto-dev/topaz/pkg/services"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func loadConfigV2(r io.Reader) (*config2.Config, error) {
	init := &config2.Config{}

	v := viper.NewWithOptions(viper.EnvKeyReplacer(newReplacer()))
	v.SetConfigType("yaml")
	// v.SetEnvPrefix("TOPAZ")
	// v.AutomaticEnv()

	v.ReadConfig(r)

	if err := v.Unmarshal(init, func(dc *mapstructure.DecoderConfig) { dc.TagName = "json" }); err != nil {
		return nil, err
	}

	return init, nil
}

func TestLoadConfigV2(t *testing.T) {
	r, err := os.Open("/Users/gertd/.config/topaz/cfg/gdrive.yaml")
	require.NoError(t, err)

	cfg2, err := loadConfigV2(r)
	require.NoError(t, err)
	require.NotNil(t, cfg2)
}

func TestMigrateConfig(t *testing.T) {
	l := zerolog.New(io.Discard)
	cfg2, err := config2.NewConfig("/Users/gertd/.config/topaz/cfg/gdrive.yaml", &l, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, cfg2)
	require.Equal(t, cfg2.Version, config2.ConfigFileVersion)

	cfg3 := &config3.Config{}

	cfg3.Version = config3.Version

	cfg3.Logging = cfg2.Logging

	cfg3.Authentication.Enabled = len(cfg2.Auth.Keys) != 0
	cfg3.Authentication.Plugin = authentication.LocalAuthenticationPlugin
	cfg3.Authentication.Settings = authentication.LocalSettings{
		Keys: cfg2.Auth.Keys,
		Options: authentication.CallOptions{
			Default: authentication.Options{
				EnableAPIKey:    cfg2.Auth.Options.Default.EnableAPIKey,
				EnableAnonymous: cfg2.Auth.Options.Default.EnableAnonymous,
			},
		},
	}

	cfg3.Authentication.Settings.Options.Overrides = []authentication.OptionOverrides{}
	for _, override2 := range cfg2.Auth.Options.Overrides {
		override3 := authentication.OptionOverrides{
			Paths:    override2.Paths,
			Override: authentication.Options(override2.Override),
		}
		cfg3.Authentication.Settings.Options.Overrides = append(cfg3.Authentication.Settings.Options.Overrides, override3)
	}

	cfg3.Debug = debug.Config{
		Enabled:         cfg2.DebugService.Enabled,
		ListenAddress:   cfg2.DebugService.ListenAddress,
		ShutdownTimeout: cfg2.DebugService.ShutdownTimeout,
	}

	cfg3.Health = health.Config{
		Enabled:       cfg2.APIConfig.Health.ListenAddress != "",
		ListenAddress: cfg2.APIConfig.Health.ListenAddress,
		Certificates:  cfg2.APIConfig.Health.Certificates,
	}

	cfg3.Metrics = metrics.Config{
		Enabled:       cfg2.APIConfig.Metrics.ListenAddress != "",
		ListenAddress: cfg2.APIConfig.Metrics.ListenAddress,
		Certificates:  cfg2.APIConfig.Metrics.Certificates,
	}

	cfg3.Services = services.Config{}

	// svcHosts := gRPC listen address -> builder.API
	svcHosts := map[string]*builder.API{}

	// port2names := gRPC listen address -> service name (includes list for v3 service definition)
	port2names := map[string][]string{}

	for name, service := range cfg2.APIConfig.Services {
		if _, ok := svcHosts[service.GRPC.ListenAddress]; !ok {
			svcHosts[service.GRPC.ListenAddress] = service
		}

		if names, ok := port2names[service.GRPC.ListenAddress]; !ok {
			port2names[service.GRPC.ListenAddress] = []string{name}
		} else {
			names = append(names, name)
			port2names[service.GRPC.ListenAddress] = names
		}
	}

	svcCounter := 0
	for addr, host := range svcHosts {
		includes := port2names[addr]

		svcCounter++
		var svc string

		switch {
		case len(svcHosts) == 1:
			svc = "topaz-svc"
		case len(includes) == 1:
			svc = fmt.Sprintf("%s-svc", includes[0])
		case lo.Contains(includes, "reader"):
			svc = "directory-svc"
		default:
			svc = fmt.Sprintf("topaz-%d-svc", svcCounter)
		}

		cfg3.Services[svc] = &services.Service{
			DependsOn: host.Needs,
			GRPC: services.GRPCService{
				ListenAddress:     host.GRPC.ListenAddress,
				FQDN:              host.GRPC.FQDN,
				Certs:             host.GRPC.Certs,
				ConnectionTimeout: time.Duration(int64(host.GRPC.ConnectionTimeoutSeconds)) * time.Second,
				DisableReflection: false,
			},
			Gateway: services.GatewayService{
				ListenAddress:     host.Gateway.ListenAddress,
				FQDN:              host.Gateway.FQDN,
				Certs:             host.Gateway.Certs,
				AllowedOrigins:    host.Gateway.AllowedOrigins,
				AllowedHeaders:    host.Gateway.AllowedHeaders,
				AllowedMethods:    host.Gateway.AllowedMethods,
				HTTP:              host.Gateway.HTTP,
				ReadTimeout:       host.Gateway.ReadTimeout,
				ReadHeaderTimeout: host.Gateway.ReadHeaderTimeout,
				WriteTimeout:      host.Gateway.WriteTimeout,
				IdleTimeout:       host.Gateway.IdleTimeout,
			},
			Includes: includes,
		}
	}

	cfg3.Directory = directory.Config{}

	// use BoltDB plugin (DEFAULT)
	if cfg2.DirectoryResolver.Address == cfg2.APIConfig.Services["reader"].GRPC.ListenAddress {
		cfg3.Directory.ReadTimeout = cfg2.Edge.RequestTimeout
		cfg3.Directory.WriteTimeout = cfg2.Edge.RequestTimeout
		cfg3.Directory.Store.Plugin = directory.BoltDBStorePlugin
		cfg3.Directory.Store.Settings = directory.BoltDBStore{Config: cfg2.Edge}.Map()
	}

	// use remote directory plugin when directory resolver address != directory gRPC reader address.
	if cfg2.DirectoryResolver.Address != cfg2.APIConfig.Services["reader"].GRPC.ListenAddress {
		cfg3.Directory.ReadTimeout = cfg2.Edge.RequestTimeout
		cfg3.Directory.WriteTimeout = cfg2.Edge.RequestTimeout
		cfg3.Directory.Store.Plugin = directory.RemoteDirectoryStorePlugin
		cfg3.Directory.Store.Settings = directory.RemoteDirectoryStore{Config: cfg2.DirectoryResolver}.Map()
	}

	cfg3.Authorizer = authorizer.Config{
		OPA: authorizer.OPAConfig{
			Config: cfg2.OPA,
		},
		DecisionLogger: authorizer.DecisionLoggerConfig{
			Plugin:   cfg2.DecisionLogger.Type,
			Settings: cfg2.DecisionLogger.Config,
		},
		Controller: authorizer.ControllerConfig{
			Config: controller.Config{
				Enabled: false,
				Server:  cfg2.ControllerConfig.Server,
			},
		},
		JWT: authorizer.JWTConfig{AcceptableTimeSkew: time.Duration(int64(cfg2.JWT.AcceptableTimeSkewSeconds)) * time.Second},
	}

	if err := cfg3.Generate(os.Stdout); err != nil {
		require.NoError(t, err)
	}
}
