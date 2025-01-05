package migrate

import (
	"fmt"
	"io"
	"os"
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
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

// Load version 2 config file, without substituting environment variables.
func LoadConfigV2(r io.Reader) (*config2.Config, error) {
	cfg2 := &config2.Config{}

	v := viper.NewWithOptions()
	v.SetConfigType("yaml")

	v.ReadConfig(r)

	if err := v.Unmarshal(cfg2, func(dc *mapstructure.DecoderConfig) { dc.TagName = "json" }); err != nil {
		return nil, err
	}

	return cfg2, nil
}

func Migrate(cfg2 *config2.Config) (*config3.Config, error) {
	cfg3 := &config3.Config{Version: config3.Version}

	cfg3.Logging = cfg2.Logging

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, "recovered in ", r)
		}
	}()

	migAuthentication(cfg2, cfg3)

	migDebug(cfg2, cfg3)

	migHealth(cfg2, cfg3)

	migMetrics(cfg2, cfg3)

	migServices(cfg2, cfg3)

	migDirectory(cfg2, cfg3)

	migAuthorizer(cfg2, cfg3)

	return cfg3, nil
}

func migAuthentication(cfg2 *config2.Config, cfg3 *config3.Config) {
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
}

func migDebug(cfg2 *config2.Config, cfg3 *config3.Config) {
	cfg3.Debug = debug.Config{
		Enabled:         cfg2.DebugService.Enabled,
		ListenAddress:   cfg2.DebugService.ListenAddress,
		ShutdownTimeout: cfg2.DebugService.ShutdownTimeout,
	}
}

func migMetrics(cfg2 *config2.Config, cfg3 *config3.Config) {
	cfg3.Metrics = metrics.Config{
		Enabled:       cfg2.APIConfig.Metrics.ListenAddress != "",
		ListenAddress: cfg2.APIConfig.Metrics.ListenAddress,
		Certificates:  cfg2.APIConfig.Metrics.Certificates,
	}
}

func migHealth(cfg2 *config2.Config, cfg3 *config3.Config) {
	cfg3.Health = health.Config{
		Enabled:       cfg2.APIConfig.Health.ListenAddress != "",
		ListenAddress: cfg2.APIConfig.Health.ListenAddress,
		Certificates:  cfg2.APIConfig.Health.Certificates,
	}
}

func migServices(cfg2 *config2.Config, cfg3 *config3.Config) {
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
}

func migDirectory(cfg2 *config2.Config, cfg3 *config3.Config) {
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
}

func migAuthorizer(cfg2 *config2.Config, cfg3 *config3.Config) {
	cfg3.Authorizer = authorizer.Config{
		OPA: authorizer.OPAConfig{
			Config: cfg2.OPA,
		},
		JWT: authorizer.JWTConfig{AcceptableTimeSkew: time.Duration(int64(cfg2.JWT.AcceptableTimeSkewSeconds)) * time.Second},
	}

	if cfg2.DecisionLogger.Type != "" && cfg2.DecisionLogger.Config != nil {
		cfg3.Authorizer.DecisionLogger = authorizer.DecisionLoggerConfig{
			Enabled:  true,
			Plugin:   cfg2.DecisionLogger.Type,
			Settings: cfg2.DecisionLogger.Config,
		}
	}

	// *ControllerConfig
	if cfg2.ControllerConfig != nil && cfg2.ControllerConfig.Enabled {
		cfg3.Authorizer.Controller = authorizer.ControllerConfig{
			Config: controller.Config{
				Enabled: cfg2.ControllerConfig.Enabled,
				Server:  cfg2.ControllerConfig.Server,
			},
		}
	}
}
