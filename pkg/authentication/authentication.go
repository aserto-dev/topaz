package authentication

import (
	"os"
	"strings"
	"text/template"

	"github.com/aserto-dev/topaz/pkg/config/handler"

	"github.com/spf13/viper"
)

// authentication:
//   enabled: false
//   plugin: local
//   settings:
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
	Enabled  bool          `json:"enabled"`
	Plugin   string        `json:"plugin,omitempty"`
	Settings LocalSettings `json:"settings,omitempty"`
}

var _ = handler.Config(&Config{})

func (c *Config) SetDefaults(v *viper.Viper, p ...string) {
	v.SetDefault(strings.Join(append(p, "enabled"), "."), false)
}

func (c *Config) Validate() (bool, error) {
	return true, nil
}

func (c *Config) Generate(w *os.File) error {
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
  plugin: {{ .Plugin }}
  settings:
    keys:
      {{- range .Settings.Keys }}
      - {{ . -}}
      {{ end }}
    options:
      default:
        enable_api_key: {{ .Settings.Options.Default.EnableAPIKey }}
        enable_anonymous: {{ .Settings.Options.Default.EnableAnonymous }}
      overrides:
        {{- range .Settings.Options.Overrides }}
        - override:
            enable_api_key: {{ .Override.EnableAPIKey }}
            enable_anonymous: {{ .Override.EnableAnonymous }}
          paths:
          {{- range .Paths }}
          - {{ . -}}
          {{ end -}}
        {{ end }}
`

// plugin: local - local authentication implementation.
type LocalSettings struct {
	Keys    []string    `json:"keys"`
	Options CallOptions `json:"options"`
}

type APIKey struct {
	Key     string `json:"key"`
	Account string `json:"account"`
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
