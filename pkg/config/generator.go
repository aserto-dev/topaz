package config

import (
	"html/template"
	"io"
	"os"
	"strings"

	"github.com/aserto-dev/topaz/internal/fs"
	"github.com/aserto-dev/topaz/topaz/cc"
	"github.com/distribution/reference"
)

const defaultPolicyRegistry string = "https://ghcr.io"

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

func (g *Generator) WithConfigName(configName string) *Generator {
	g.ConfigName = configName
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
	g.PolicyRegistry = defaultPolicyRegistry // set to original default

	policyRegistry, _, found := strings.Cut(resource, "/")
	if found && policyRegistry != "" {
		g.PolicyRegistry = "https://" + policyRegistry
	}

	g.Resource = resource

	ref, err := reference.ParseDockerRef(g.Resource)
	if err == nil {
		g.RegistryService = reference.Domain(ref)
		g.RegistryImage = reference.Path(ref)

		if tagged, ok := ref.(reference.Tagged); ok {
			g.RegistryTag = tagged.Tag()
		}
	}

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

func (g *Generator) GenerateConfig(w io.Writer, templateData string) error {
	return g.writeConfig(w, templateData)
}

func (g *Generator) CreateConfigDir() (string, error) {
	configDir := cc.GetTopazCfgDir()
	if fi, err := os.Stat(configDir); err == nil && fi.IsDir() {
		return configDir, nil
	}

	return configDir, os.MkdirAll(configDir, fs.FileModeOwnerRW)
}

func (g *Generator) CreateCertsDir() (string, error) {
	certsDir := cc.GetTopazCertsDir()
	if fi, err := os.Stat(certsDir); err == nil && fi.IsDir() {
		return certsDir, nil
	}

	return certsDir, os.MkdirAll(certsDir, fs.FileModeOwnerRW)
}

func (g *Generator) CreateDataDir() (string, error) {
	dataDir := cc.GetTopazDataDir()
	if fi, err := os.Stat(dataDir); err == nil && fi.IsDir() {
		return dataDir, nil
	}

	return dataDir, os.MkdirAll(dataDir, fs.FileModeOwnerRW)
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
