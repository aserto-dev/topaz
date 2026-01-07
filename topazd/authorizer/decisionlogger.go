package authorizer

import (
	"io"
	"iter"
	"text/template"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/topazd/loiter"
)

const (
	pluginIndentLevel = 2
)

type DecisionLoggerConfig struct {
	config.Optional

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

func (c *DecisionLoggerConfig) Paths() iter.Seq2[string, config.AccessMode] {
	if c.Enabled {
		switch c.Provider {
		case FileDecisionLoggerPlugin:
			return c.File.Paths()
		case SelfDecisionLoggerPlugin:
			return c.Self.Paths()
		}
	}

	return loiter.Seq2[string, config.AccessMode]()
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

const decisionLoggerConfigTemplate = `# decision logger configuration.
decision_logger:
  enabled: {{ .Enabled }}
  {{- if .Provider }}
  provider: {{ .Provider }}
  {{- end }}
`
