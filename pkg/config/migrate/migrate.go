package migrate

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aserto-dev/self-decision-logger/logger/self"
	"github.com/aserto-dev/topaz/controller"
	"github.com/aserto-dev/topaz/decisionlog/logger/file"
	"github.com/aserto-dev/topaz/pkg/authentication"
	"github.com/aserto-dev/topaz/pkg/authorizer"
	config2 "github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/debug"
	"github.com/aserto-dev/topaz/pkg/directory"
	"github.com/aserto-dev/topaz/pkg/health"
	"github.com/aserto-dev/topaz/pkg/metrics"
	"github.com/aserto-dev/topaz/pkg/servers"
	"github.com/aserto-dev/topaz/pkg/service/builder"
	config3 "github.com/aserto-dev/topaz/pkg/topaz"
	"github.com/go-viper/mapstructure/v2"
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
	cfg3.Authentication = authentication.Config{
		Enabled:  len(cfg2.Auth.Keys) != 0,
		Provider: authentication.LocalAuthenticationPlugin,
		Local: authentication.LocalConfig{
			Keys: cfg2.Auth.Keys,
			Options: authentication.CallOptions{
				Default: authentication.Options{
					EnableAPIKey:    cfg2.Auth.Options.Default.EnableAPIKey,
					EnableAnonymous: cfg2.Auth.Options.Default.EnableAnonymous,
				},
				Overrides: lo.Map(
					cfg2.Auth.Options.Overrides,
					func(override2 config2.OptionOverrides, _ int) authentication.OptionOverrides {
						return authentication.OptionOverrides{
							Paths:    override2.Paths,
							Override: authentication.Options(override2.Override),
						}
					},
				),
			},
		},
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
	cfg3.Servers = servers.Config{}

	// svcHosts := gRPC listen address -> builder.API
	svcHosts := map[string]*builder.API{}

	// port2names := gRPC listen address -> service name (includes list for v3 service definition)
	port2names := map[string][]servers.ServiceName{}

	for name, service := range cfg2.APIConfig.Services {
		svcHosts[service.GRPC.ListenAddress] = service
		port2names[service.GRPC.ListenAddress] = append(port2names[service.GRPC.ListenAddress], servers.ServiceName(name))
	}

	svcCounter := 0

	for addr, host := range svcHosts {
		includes := port2names[addr]

		svcCounter++

		var svc servers.ServerName

		switch {
		case len(svcHosts) == 1:
			svc = "topaz-svc"
		case len(includes) == 1:
			svc = servers.ServerName(includes[0] + "-svc")
		case lo.Contains(includes, "reader"):
			svc = "directory-svc"
		default:
			svc = servers.ServerName(fmt.Sprintf("topaz-%d-svc", svcCounter))
		}

		cfg3.Servers[svc] = &servers.Server{
			DependsOn: lo.Map(host.Needs, func(name string, _ int) servers.ServerName { return servers.ServerName(name) }),
			GRPC: servers.GRPCServer{
				ListenAddress:     host.GRPC.ListenAddress,
				FQDN:              host.GRPC.FQDN,
				Certs:             host.GRPC.Certs,
				ConnectionTimeout: time.Duration(int64(host.GRPC.ConnectionTimeoutSeconds)) * time.Second,
				DisableReflection: false,
			},
			HTTP: servers.HTTPServer{
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
			Services: includes,
		}
	}
}

func migDirectory(cfg2 *config2.Config, cfg3 *config3.Config) {
	// when directory resolver address == directory gRPC reader address
	// use BoltDB plugin (DEFAULT)
	if cfg2.DirectoryResolver.Address == cfg2.APIConfig.Services["reader"].GRPC.ListenAddress {
		cfg3.Directory = directory.Config{
			ReadTimeout:  cfg2.Edge.RequestTimeout,
			WriteTimeout: cfg2.Edge.RequestTimeout,
			Store: directory.Store{
				Provider: directory.BoltDBStorePlugin,
				Bolt:     directory.BoltDBStore(cfg2.Edge),
			},
		}
	} else {
		cfg3.Directory = directory.Config{
			ReadTimeout:  cfg2.Edge.RequestTimeout,
			WriteTimeout: cfg2.Edge.RequestTimeout,
			Store: directory.Store{
				Provider: directory.RemoteDirectoryStorePlugin,
				Remote:   directory.RemoteDirectoryStore(cfg2.DirectoryResolver),
			},
		}
	}
}

func migAuthorizer(cfg2 *config2.Config, cfg3 *config3.Config) {
	cfg3.Authorizer = authorizer.Config{
		OPA: authorizer.OPAConfig(cfg2.OPA),
		JWT: authorizer.JWTConfig{AcceptableTimeSkew: time.Duration(int64(cfg2.JWT.AcceptableTimeSkewSeconds)) * time.Second},
	}

	if cfg2.DecisionLogger.Type != "" && len(cfg2.DecisionLogger.Config) > 0 {
		cfg3.Authorizer.DecisionLogger = migDecisionLogger(&cfg2.DecisionLogger)
	}

	// *ControllerConfig
	if cfg2.ControllerConfig.Enabled {
		cfg3.Authorizer.Controller = authorizer.ControllerConfig(
			controller.Config{
				Enabled: cfg2.ControllerConfig.Enabled,
				Server:  cfg2.ControllerConfig.Server,
			},
		)
	}
}

func migDecisionLogger(cfg2 *config2.DecisionLogConfig) authorizer.DecisionLoggerConfig {
	return authorizer.DecisionLoggerConfig{
		Enabled:  true,
		Provider: cfg2.Type,
		File:     migFileLogger(cfg2),
		Self:     migSelfLogger(cfg2),
	}
}

func migFileLogger(cfg2 *config2.DecisionLogConfig) authorizer.FileDecisionLoggerConfig {
	cfg3 := authorizer.FileDecisionLoggerConfig{}

	if cfg2.Type != authorizer.FileDecisionLoggerPlugin {
		return cfg3
	}

	var fileConfig file.Config

	cfgMap := lo.Assign(cfg3.Defaults(), cfg2.Config)
	if err := decodeMap(cfgMap, &fileConfig); err != nil {
		return cfg3
	}

	return authorizer.FileDecisionLoggerConfig(fileConfig)
}

func migSelfLogger(cfg2 *config2.DecisionLogConfig) authorizer.SelfDecisionLoggerConfig {
	cfg3 := authorizer.SelfDecisionLoggerConfig{}

	if cfg2.Type != authorizer.SelfDecisionLoggerPlugin {
		return cfg3
	}

	var selfConfig self.Config

	cfgMap := lo.Assign(cfg3.Defaults(), cfg2.Config)
	if err := decodeMap(cfgMap, &selfConfig); err != nil {
		return cfg3
	}

	return authorizer.SelfDecisionLoggerConfig(selfConfig)
}

func decodeMap[T any](m map[string]any, c *T) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  c,
		TagName: "json",
	})
	if err != nil {
		return err
	}

	return decoder.Decode(m)
}
