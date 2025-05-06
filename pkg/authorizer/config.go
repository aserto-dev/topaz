package authorizer

import (
	"io"
	"text/template"

	"github.com/aserto-dev/topaz/pkg/config"
)

type Config struct {
	OPA            OPAConfig            `json:"opa"`
	DecisionLogger DecisionLoggerConfig `json:"decision_logger"`
	Controller     ControllerConfig     `json:"controller"`
	JWT            JWTConfig            `json:"jwt"`
}

var _ config.Section = (*Config)(nil)

func (c *Config) Defaults() map[string]any {
	return config.PrefixKeys("jwt", c.JWT.Defaults())
}

func (c *Config) Validate() error {
	return nil
}

func (c *Config) Serialize(w io.Writer) error {
	tmpl, err := template.New("AUTHORIZER").Parse(config.TrimN(configTemplate))
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	w = config.IndentWriter(w, sectionIndentLevel)

	if err := c.OPA.Serialize(w); err != nil {
		return err
	}

	if err := c.DecisionLogger.Serialize(w); err != nil {
		return err
	}

	if err := c.Controller.Serialize(w); err != nil {
		return err
	}

	if err := c.JWT.Serialize(w); err != nil {
		return err
	}

	return nil
}

const (
	sectionIndentLevel = 2

	configTemplate = `
# authorizer configuration.
authorizer:
`
)
