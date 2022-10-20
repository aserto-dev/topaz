package cmd

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"path"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
)

type ConfigureCmd struct {
	PolicyName    string `arg:"" required:"" help:"policy name"`
	Resource      string `short:"r" required:"" help:"resource url"`
	Stdout        bool   `short:"p" help:"generated configuration is printed to stdout but not saved"`
	EdgeDirectory bool   `short:"d" help:"enable edge directory" default:"false"`
	SeedMetadata  bool   `short:"s" help:"enable seed metadata" default:"false"`
}

func (cmd ConfigureCmd) Run(c *cc.CommonCtx) error {
	fmt.Fprintf(c.UI.Err(), ">>> configure policy...\n")

	configDir, err := CreateConfigDir()
	if err != nil {
		return err
	}

	params := templateParams{
		PolicyName:    cmd.PolicyName,
		Resource:      cmd.Resource,
		EdgeDirectory: cmd.EdgeDirectory,
		SeedMetadata:  cmd.SeedMetadata,
	}

	fmt.Fprintf(c.UI.Err(), "policy name: %s\n", params.PolicyName)

	var w io.Writer

	if cmd.Stdout {
		w = c.UI.Output()
	} else {
		w, err = os.Create(path.Join(configDir, "config.yaml"))
		if err != nil {
			return err
		}
	}

	return WriteConfig(w, configTemplate, &params)
}

func CreateConfigDir() (string, error) {
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

func WriteConfig(w io.Writer, templ string, params *templateParams) error {
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
