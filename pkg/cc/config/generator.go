package config

import (
	"html/template"
	"io"
	"os"
	"path"
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
func (g *Generator) WithLocalPolicyImage(image string) *Generator {
	g.LocalPolicyImage = image
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
	g.EnableDirectoryV2 = enabled
	return g
}

func (g *Generator) GenerateConfig(w io.Writer, templateData string) error {
	return g.writeConfig(w, templateData)
}

func (g *Generator) CreateConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := path.Join(home, "/.config/topaz/cfg")
	if fi, err := os.Stat(configDir); err == nil && fi.IsDir() {
		return configDir, nil
	}

	return configDir, os.MkdirAll(configDir, 0700)
}

func (g *Generator) CreateCertsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	certsDir := path.Join(home, "/.config/topaz/certs")
	if fi, err := os.Stat(certsDir); err == nil && fi.IsDir() {
		return certsDir, nil
	}

	return certsDir, os.MkdirAll(certsDir, 0700)
}

func (g *Generator) CreateDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dataDir := path.Join(home, "/.config/topaz/db")
	if fi, err := os.Stat(dataDir); err == nil && fi.IsDir() {
		return dataDir, nil
	}

	return dataDir, os.MkdirAll(dataDir, 0700)
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
