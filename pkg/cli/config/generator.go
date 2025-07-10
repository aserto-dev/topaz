package config

import (
	"bytes"
	"io"
	"path"
	"strings"

	"github.com/open-policy-agent/opa/v1/download"
	"github.com/open-policy-agent/opa/v1/plugins/bundle"
	"github.com/samber/lo"

	"github.com/aserto-dev/topaz/pkg/cli/x"
	cfgutil "github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/pkg/config/v3"
	"github.com/aserto-dev/topaz/pkg/servers"
)

const (
	DefaultPolicyRegistry               = "https://ghcr.io"
	PolicyRegistryServiceName           = "policy-registry"
	PolicyRegistryResponseHeaderTimeout = 5 // seconds

	BundlePollingMinDelay int64 = 60  // seconds
	BundlePollingMaxDelay int64 = 120 // seconds
)

type Generator struct {
	name string
	cfg  *config.Config
}

type GeneratorOption func(name string, cfg *config.Config)

func NewGenerator(name string, opts ...GeneratorOption) (*Generator, error) {
	cfg, err := config.NewConfig(&bytes.Buffer{}, defaultServer, boltdbFile(name))
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(name, cfg)
	}

	return &Generator{
		name: name,
		cfg:  cfg,
	}, nil
}

func WithBundle(resource string) GeneratorOption {
	return func(name string, cfg *config.Config) {
		setPolicyRegistry(resource, cfg)

		bundles := cfg.Authorizer.OPA.Config.Bundles
		cfg.Authorizer.OPA.Config.Bundles = lo.Assign(bundles, map[string]*bundle.Source{
			name: {
				Config: download.Config{
					Polling: download.PollingConfig{
						MinDelaySeconds: Ptr(BundlePollingMinDelay),
						MaxDelaySeconds: Ptr(BundlePollingMaxDelay),
					},
				},
				Service:  PolicyRegistryServiceName,
				Resource: resource,
				Persist:  false,
			},
		})
	}
}

func WithLocalBundle(path string) GeneratorOption {
	return func(name string, cfg *config.Config) {
		cfg.Authorizer.OPA.LocalBundles.LocalPolicyImage = path
		cfg.Authorizer.OPA.LocalBundles.Watch = true
		cfg.Authorizer.OPA.LocalBundles.SkipVerification = true
	}
}

func (g *Generator) Generate(w io.Writer) error {
	return g.cfg.Serialize(w)
}

func setPolicyRegistry(resource string, cfg *config.Config) {
	var registry string

	if policyRegistry, _, found := strings.Cut(resource, "/"); found && policyRegistry != "" {
		registry = "https://" + policyRegistry
	} else {
		registry = DefaultPolicyRegistry
	}

	// add policy registry to any existing services that may have been set.
	services := cfg.Authorizer.OPA.Config.Services
	cfg.Authorizer.OPA.Config.Services = lo.Assign(services, map[string]any{
		PolicyRegistryServiceName: map[string]any{
			"url":                             registry,
			"type":                            "oci",
			"response_header_timeout_seconds": PolicyRegistryResponseHeaderTimeout,
		},
	})
}

// defaultServer is a config override that defines a single 'topaz' server on ports 8282 and 8383 running all services.
func defaultServer(cfg *config.Config) {
	var s servers.Server

	v := cfgutil.NewViper()
	v.SetDefaults(&s)
	v.ReadConfig(&bytes.Buffer{})

	if err := v.UnmarshalExact(&s, cfgutil.UseJSONTags, cfgutil.WithSquash); err != nil {
		// should never happen.
		panic(err)
	}

	s.Services = servers.KnownServices

	cfg.Servers = map[servers.ServerName]*servers.Server{
		"topaz": &s,
	}
}

func boltdbFile(name string) config.ConfigOverride {
	return func(cfg *config.Config) {
		cfg.Directory.Store.Bolt.DBPath = path.Join("${"+x.EnvTopazDBDir+"}", name+".db")
	}
}

func Ptr[T any](v T) *T {
	return &v
}
