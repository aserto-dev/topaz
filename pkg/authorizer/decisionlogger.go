package authorizer

import (
	"io"
	"text/template"

	"github.com/aserto-dev/topaz/pkg/config"
)

const (
	pluginIndentLevel = 2
)

type DecisionLoggerConfig struct {
	Enabled  bool                     `json:"enabled"`
	Provider string                   `json:"provider"`
	File     FileDecisionLoggerConfig `json:"file"`
	Self     SelfDecisionLoggerConfig `json:"self"`
}

var _ config.Section = (*DecisionLoggerConfig)(nil)

func (c *DecisionLoggerConfig) Defaults() map[string]any {
	return map[string]any{}
}

func (c *DecisionLoggerConfig) Validate() error {
	return nil
}

func (c *DecisionLoggerConfig) Serialize(w io.Writer) error {
	tmpl, err := template.New("DECISION_LOGGER").Parse(decisionLoggerConfigTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	return c.generatePlugins(config.IndentWriter(w, pluginIndentLevel))
}

func (c *DecisionLoggerConfig) generatePlugins(w io.Writer) error {
	if err := config.WriteNonEmpty(w, &c.File); err != nil {
		return err
	}

	if err := config.WriteNonEmpty(w, &c.Self); err != nil {
		return err
	}

	return nil
}

const decisionLoggerConfigTemplate = `
# decision logger configuration.
decision_logger:
  enabled: {{ .Enabled }}
  {{- if .Provider }}
  provider: {{ .Provider }}
  {{- end }}
`
