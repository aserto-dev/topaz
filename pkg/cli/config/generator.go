package config

import (
	"bytes"
	"io"
	"strings"

	"github.com/open-policy-agent/opa/v1/download"
	"github.com/open-policy-agent/opa/v1/plugins/bundle"
	"github.com/samber/lo"

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
	ConfigName string

	cfg *config.Config
}

func NewGenerator(configName string) (*Generator, error) {
	cfg, err := config.NewConfig(&bytes.Buffer{}, defaultServer)
	if err != nil {
		return nil, err
	}

	return &Generator{
		ConfigName: configName,
		cfg:        cfg,
	}, nil
}

func (g *Generator) WithBundle(name, resource string) *Generator {
	g.setPolicyRegistry(resource)

	bundles := g.cfg.Authorizer.OPA.Config.Bundles
	g.cfg.Authorizer.OPA.Config.Bundles = lo.Assign(bundles, map[string]*bundle.Source{
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

	return g
}

func (g *Generator) WithLocalBundle(path string) *Generator {
	g.cfg.Authorizer.OPA.LocalBundles.LocalPolicyImage = path
	g.cfg.Authorizer.OPA.LocalBundles.Watch = true
	g.cfg.Authorizer.OPA.LocalBundles.SkipVerification = true

	return g
}

func (g *Generator) Generate(w io.Writer) error {
	return g.cfg.Serialize(w)
}

func (g *Generator) setPolicyRegistry(resource string) {
	var registry string

	if policyRegistry, _, found := strings.Cut(resource, "/"); found && policyRegistry != "" {
		registry = "https://" + policyRegistry
	} else {
		registry = DefaultPolicyRegistry
	}

	// add policy registry to any existing services that may have been set.
	services := g.cfg.Authorizer.OPA.Config.Services
	g.cfg.Authorizer.OPA.Config.Services = lo.Assign(services, map[string]any{
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

func Ptr[T any](v T) *T {
	return &v
}
