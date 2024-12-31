package authorizer

import (
	"os"
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

func (c *Config) Generate(w *os.File) error {
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

// func (c *Config) OPA() (*runtime.Config, error) {
// 	rCfg := &runtime.Config{}

// 	b, err := json.Marshal(c.RawOPA)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if err := json.Unmarshal(b, rCfg); err != nil {
// 		return nil, err
// 	}

// 	return rCfg, nil
// }

// func (c *Config) UnmarshalJSON(data []byte) error {
// 	rCfg := runtime.Config{}

// 	b, err := json.Marshal(c.RawOPA)
// 	if err != nil {
// 		return err
// 	}

// 	if err := json.Unmarshal(b, &rCfg); err != nil {
// 		return err
// 	}

// 	c.OPA = rCfg

// 	return nil
// }

// func (c *Config) MarshalJSON() ([]byte, error) {
// 	return []byte{}, nil
// }
