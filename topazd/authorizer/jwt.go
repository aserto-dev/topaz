package authorizer

import (
	"io"
	"iter"
	"text/template"
	"time"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/topazd/loiter"
)

const DefaultAcceptableTimeSkew = time.Second * 5

type JWTConfig struct {
	AcceptableTimeSkew time.Duration `json:"acceptable_time_skew"`
}

var _ config.Section = (*JWTConfig)(nil)

func (c *JWTConfig) Defaults() map[string]any {
	return map[string]any{
		"acceptable_time_skew": DefaultAcceptableTimeSkew.String(),
	}
}

func (c *JWTConfig) Validate() error {
	return nil
}

func (c *JWTConfig) Paths() iter.Seq2[string, config.AccessMode] {
	return loiter.Seq2[string, config.AccessMode]()
}

func (c *JWTConfig) Serialize(w io.Writer) error {
	tmpl, err := template.New("JWT").Parse(jwtConfigTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	return nil
}

const jwtConfigTemplate = `# jwt validation configuration
jwt:
  acceptable_time_skew: {{ .AcceptableTimeSkew -}}
`
