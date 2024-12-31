package authorizer

import (
	"os"
	"text/template"
	"time"

	"github.com/spf13/viper"
)

const DefaultAcceptableTimeSkew = time.Second * 5

type JWTConfig struct {
	AcceptableTimeSkew time.Duration `json:"acceptable_time_skew"`
}

func (c *JWTConfig) SetDefaults(v *viper.Viper, p ...string) {
	v.SetDefault("acceptable_time_skew", DefaultAcceptableTimeSkew.String())
}

func (c *JWTConfig) Validate() (bool, error) {
	return true, nil
}

func (c *JWTConfig) Generate(w *os.File) error {
	tmpl, err := template.New("JWT").Parse(jwtConfigTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	return nil
}

const jwtConfigTemplate = `
  # jwt validation configuration
  jwt:
    acceptable_time_skew: {{ .AcceptableTimeSkew }}
`
