package config

import (
	"html/template"
	"io"
	"os"
	"strings"

	"github.com/open-policy-agent/opa/v1/download"
	"github.com/open-policy-agent/opa/v1/plugins/bundle"
	"github.com/samber/lo"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/x"
	"github.com/aserto-dev/topaz/pkg/config/v3"
)

const (
	DefaultPolicyRegistry               = "https://ghcr.io"
	PolicyRegistryServiceName           = "policy-registry"
	PolicyRegistryResponseHeaderTimeout = 5 // seconds

	BundlePollingMinDelay int64 = 60  // seconds
	BundlePollingMaxDelay int64 = 120 // seconds
)

type Generator struct {
	templateParams
	ConfigName string

	cfg config.Config
}

func NewGenerator(configName string) *Generator {
	return &Generator{
		ConfigName: configName,
		cfg:        config.Config{Version: config.Version},
	}
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

func (g *Generator) WithTenantID(tenantID string) *Generator {
	g.TenantID = tenantID
	return g
}

func (g *Generator) WithDiscovery(url, key string) *Generator {
	g.DiscoveryURL = url
	g.TenantKey = key

	return g
}

func (g *Generator) WithController(url, clientCertPath, clientKeyPath string) *Generator {
	g.ControlPlane.Enabled = true
	g.ControlPlane.Address = url
	g.ControlPlane.ClientCertPath = clientCertPath
	g.ControlPlane.ClientKeyPath = clientKeyPath

	return g
}

func (g *Generator) WithSelfDecisionLogger(emsURL, clientCertPath, clientKeyPath, storePath string) *Generator {
	g.DecisionLogging = true
	g.DecisionLogger.EMSAddress = emsURL
	g.DecisionLogger.ClientCertPath = clientCertPath
	g.DecisionLogger.ClientKeyPath = clientKeyPath
	g.DecisionLogger.StorePath = storePath

	return g
}

func (g *Generator) GenerateConfig(w io.Writer, templateData string) error {
	return g.writeConfig(w, templateData)
}

func (g *Generator) CreateConfigDir() (string, error) {
	configDir := cc.GetTopazCfgDir()
	if fi, err := os.Stat(configDir); err == nil && fi.IsDir() {
		return configDir, nil
	}

	return configDir, os.MkdirAll(configDir, x.FileMode0700)
}

func (g *Generator) CreateCertsDir() (string, error) {
	certsDir := cc.GetTopazCertsDir()
	if fi, err := os.Stat(certsDir); err == nil && fi.IsDir() {
		return certsDir, nil
	}

	return certsDir, os.MkdirAll(certsDir, x.FileMode0700)
}

func (g *Generator) CreateDataDir() (string, error) {
	dataDir := cc.GetTopazDataDir()
	if fi, err := os.Stat(dataDir); err == nil && fi.IsDir() {
		return dataDir, nil
	}

	return dataDir, os.MkdirAll(dataDir, x.FileMode0700)
}

func (g *Generator) writeConfig(w io.Writer, templ string) error {
	t, err := template.New("config").Parse(templ)
	if err != nil {
		return err
	}

	err = t.Execute(w, g)
	if err != nil {
		return err
	}

	return nil
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

func Ptr[T any](v T) *T {
	return &v
}
