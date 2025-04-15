package authorizer

import (
	"bytes"
	"io"
	"text/template"

	"github.com/aserto-dev/topaz/pkg/config"
)

const (
	pluginIndentLevel = 2
)

type DecisionLoggerConfig struct {
	Enabled bool                     `json:"enabled"`
	Use     string                   `json:"use"`
	File    FileDecisionLoggerConfig `json:"file"`
	Self    SelfDecisionLoggerConfig `json:"self"`
}

var _ config.Section = (*DecisionLoggerConfig)(nil)

func (c *DecisionLoggerConfig) Defaults() map[string]any {
	return map[string]any{}
}

func (c *DecisionLoggerConfig) Validate() (bool, error) {
	return true, nil
}

func (c *DecisionLoggerConfig) Generate(w io.Writer) error {
	tmpl, err := template.New("DECISION_LOGGER").Parse(decisionLoggerConfigTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := c.generatePlugins(&buf); err != nil {
		return err
	}

	_, err = w.Write([]byte(config.Indent(buf.String(), pluginIndentLevel)))

	return err
}

func (c *DecisionLoggerConfig) generatePlugins(w io.Writer) error {
	if err := config.WriteIfNotEmpty(w, &c.File); err != nil {
		return err
	}

	if err := config.WriteIfNotEmpty(w, &c.Self); err != nil {
		return err
	}

	return nil
}

const decisionLoggerConfigTemplate = `
# decision logger configuration.
decision_logger:
  enabled: {{ .Enabled }}
  {{- if .Use }}
  use: {{ .Use }}
  {{- end }}
`
