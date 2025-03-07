package config

import (
	"html/template"
	"io"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/fs"
)

type Generator struct {
	templateParams
	ConfigName string
}

func NewGenerator(configName string) *Generator {
	return &Generator{ConfigName: configName}
}

func (g *Generator) WithVersion(version int) *Generator {
	g.Version = version
	return g
}

func (g *Generator) WithLocalPolicy(local bool) *Generator {
	g.LocalPolicy = local
	return g
}

func (g *Generator) WithPolicyName(policyName string) *Generator {
	g.PolicyName = policyName
	return g
}

func (g *Generator) WithResource(resource string) *Generator {
	g.Resource = resource
	return g
}

func (g *Generator) WithEdgeDirectory(enabled bool) *Generator {
	g.EdgeDirectory = enabled
	return g
}

func (g *Generator) WithEnableDirectoryV2(enabled bool) *Generator {
	g.EnableDirectoryV2 = false
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
	if fs.DirExists(configDir) {
		return configDir, nil
	}

	return configDir, fs.EnsureDirPath(configDir, fs.FileMode0700)
}

func (g *Generator) CreateCertsDir() (string, error) {
	certsDir := cc.GetTopazCertsDir()
	if fs.DirExists(certsDir) {
		return certsDir, nil
	}

	return certsDir, fs.EnsureDirPath(certsDir, fs.FileMode0755)
}

func (g *Generator) CreateDataDir() (string, error) {
	dataDir := cc.GetTopazDataDir()
	if fs.DirExists(dataDir) {
		return dataDir, nil
	}

	return dataDir, fs.EnsureDirPath(dataDir, fs.FileMode0700)
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
