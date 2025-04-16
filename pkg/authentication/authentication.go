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

func (c *Config) Defaults() map[string]any {
	return map[string]any{
		"enabled": false,
	}
}

func (c *Config) Validate() error {
	return nil
}

func (c *Config) Serialize(w io.Writer) error {
	tmpl, err := template.New("AUTHENTICATION").Parse(authenticationTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	return nil
}

const LocalAuthenticationPlugin string = "local"

const authenticationTemplate = `
# local authentication configuration.
authentication:
  enabled: {{ .Enabled }}
  provider: {{ .Provider }}
  local:
    keys:
      {{- range .Local.Keys }}
      - {{ . -}}
      {{ end }}
    options:
      default:
        enable_api_key: {{ .Local.Options.Default.EnableAPIKey }}
        enable_anonymous: {{ .Local.Options.Default.EnableAnonymous }}
      overrides:
        {{- range .Local.Options.Overrides }}
        - override:
            enable_api_key: {{ .Override.EnableAPIKey }}
            enable_anonymous: {{ .Override.EnableAnonymous }}
          paths:
          {{- range .Paths }}
          - {{ . -}}
          {{ end -}}
        {{ end }}
`

// provider: local - local authentication implementation.
type LocalConfig struct {
	Keys    []string    `json:"keys"`
	Options CallOptions `json:"options"`
}

type CallOptions struct {
	Default   Options           `json:"default"`
	Overrides []OptionOverrides `json:"overrides"`
}

type Options struct {
	// API Key for machine-to-machine communication, internal to Aserto
	EnableAPIKey bool `json:"enable_api_key"`
	// Allows calls without any form of authentication
	EnableAnonymous bool `json:"enable_anonymous"`
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
