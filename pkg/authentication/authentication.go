package authentication

import (
	"io"
	"strings"
	"text/template"

	"github.com/aserto-dev/topaz/pkg/config"
)

// authentication:
//   enabled: false
//   use: local
//   local:
//     keys:
//       - "69ba614c64ed4be69485de73d062a00b"
//       - "##Ve@rySecret123!!"
//     options:
//       default:
//         enable_api_key: true
//         enable_anonymous: false
//       overrides:
//         paths:
//           - /aserto.authorizer.v2.Authorizer/Info
//           - /grpc.reflection.v1.ServerReflection/ServerReflectionInfo
//           - /grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo
//         override:
//           enable_anonymous: true
//           enable_api_key: false

type Config struct {
	Enabled  bool        `json:"enabled"`
	Provider string      `json:"provider,omitempty"`
	Local    LocalConfig `json:"local,omitempty"`
}

var _ config.Section = (*Config)(nil)

func (*Config) Defaults() map[string]any {
	return map[string]any{
		"enabled": false,
	}
}

func (*Config) Validate() error {
	return nil
}

func (c *Config) Serialize(w io.Writer) error {
	tmpl, err := template.New("AUTHENTICATION").
		Funcs(config.TemplateFuncs()).
		Parse(authenticationTemplate)
	if err != nil {
		return err
	}

	return tmpl.Execute(w, c)
}

func (c *Config) ReaderKey() string {
	key := ""

	if c.Enabled && len(c.Local.Keys) > 0 {
		key = c.Local.Keys[0]
	}

	return key
}

const LocalAuthenticationPlugin string = "local"

const authenticationTemplate = `
# local authentication configuration.
authentication:
  enabled: {{ .Enabled }}
  provider: {{ .Provider }}
  local:
  {{- with .Local }}
    keys:
      {{- .Keys | toYAML | nindent 6 }}
    options:
      default:
        allow_anonymous: {{ .Options.Default.AllowAnonymous }}

    {{- with .Options.Overrides }}

      overrides:
      {{- range . }}
        - override:
            allow_anonymous: {{ .Override.AllowAnonymous }}
          paths:
            {{- .Paths | toYAML | nindent 12 }}
      {{- end }}
    {{- end }}
  {{- end }}

`

// LocalConfig holds configuration options for local authentication implementation.
type LocalConfig struct {
	Keys    []string    `json:"keys"`
	Options CallOptions `json:"options"`
}

type CallOptions struct {
	Default   Options           `json:"default"`
	Overrides []OptionOverrides `json:"overrides"`
}

type Options struct {
	// Allows calls without any form of authentication
	AllowAnonymous bool `json:"allow_anonymous"`
}

type OptionOverrides struct {
	// API paths to override
	Paths []string `json:"paths"`
	// Override options
	Override Options `json:"override"`
}

func (co *CallOptions) ForPath(path string) *Options {
	for _, override := range co.Overrides {
		for _, prefix := range override.Paths {
			if strings.HasPrefix(strings.ToLower(path), prefix) {
				return &override.Override
			}
		}
	}

	return &co.Default
}
