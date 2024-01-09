package configuration

import (
	"html/template"
	"io"
	"os"
	"path"
)

type Generator struct {
	ConfigName string
}

func NewGenerator(configName string) *Generator {
	return &Generator{ConfigName: configName}
}

func (g *Generator) GenerateConfig(w io.Writer, templateData string, params *TemplateParams) error {
	return writeConfig(w, templateData, params)
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

func writeConfig(w io.Writer, templ string, params *TemplateParams) error {
	t, err := template.New("config").Parse(templ)
	if err != nil {
		return err
	}

	err = t.Execute(w, params)
	if err != nil {
		return err
	}

	return nil
}
