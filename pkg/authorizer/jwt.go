package authorizer

import (
	"io"
	"text/template"
	"time"
)

const DefaultAcceptableTimeSkew = time.Second * 5

type JWTConfig struct {
	AcceptableTimeSkew time.Duration `json:"acceptable_time_skew"`
}

func (c *JWTConfig) Defaults() map[string]any {
	return map[string]any{
		"acceptable_time_skew": DefaultAcceptableTimeSkew.String(),
	}
}

func (c *JWTConfig) Validate() (bool, error) {
	return true, nil
}

func (c *JWTConfig) Generate(w io.Writer) error {
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
