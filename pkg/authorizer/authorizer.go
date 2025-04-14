package authorizer

import (
	"io"
	"text/template"

	"github.com/aserto-dev/topaz/pkg/config/handler"

	"github.com/spf13/viper"
)

type Config struct {
	RawOPA         map[string]interface{} `json:"opa"`
	OPA            OPAConfig              `json:"-,"`
	DecisionLogger DecisionLoggerConfig   `json:"decision_logger"`
	Controller     ControllerConfig       `json:"controller"`
	JWT            JWTConfig              `json:"jwt"`
}

var _ = handler.Config(&Config{})

func (c *Config) SetDefaults(v *viper.Viper, p ...string) {
	c.JWT.SetDefaults(v)
}

func (c *Config) Validate() (bool, error) {
	return true, nil
}

func (c *Config) Generate(w io.Writer) error {
	tmpl, err := template.New("AUTHORIZER").Parse(authorizerTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	if err := c.OPA.Generate(w); err != nil {
		return err
	}

	if err := c.DecisionLogger.Generate(w); err != nil {
		return err
	}

	if err := c.Controller.Generate(w); err != nil {
		return err
	}

	if err := c.JWT.Generate(w); err != nil {
		return err
	}

	return nil
}

const authorizerTemplate = `
# authorizer configuration.
authorizer:
`
