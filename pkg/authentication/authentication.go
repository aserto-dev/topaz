package authentication

import (
	"io"
	"iter"
	"strings"
	"text/template"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/pkg/loiter"
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
	config.Optional

	Provider string      `json:"provider,omitempty"`
	Local    LocalConfig `json:"local,omitempty"`
}

var _ config.Section = (*Config)(nil)

func (*Config) Defaults() map[string]any {
	return map[string]any{
		"enabled":  false,
		"provider": "local",
		"local": map[string]any{
			"keys": []string{},
			"options": map[string]any{
				"default": map[string]any{
					"allow_anonymous": false,
				},
				"overrides": []map[string]any{
					{
						"paths": []string{
							"/grpc.reflection.v1.ServerReflection/ServerReflectionInfo",
							"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
						},
						"override": map[string]any{
							"allow_anonymous": true,
						},
					},
				},
			},
		},
	}
}

func (*Config) Validate() error {
	return nil
}

func (*Config) Paths() iter.Seq2[string, config.AccessMode] {
	return loiter.Seq2[string, config.AccessMode]()
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
    keys: {{- if not .Keys }} [] {{- else }}
      {{- .Keys | toYAML | nindent 6 }}
	{{- end }}

	{{- with .TryOptions }}
    options:
      {{- . | toIndentYAML 2 | nindent 6 }}
    {{- end }}
  {{- end }}

`

// LocalConfig holds configuration options for local authentication implementation.
type LocalConfig struct {
	Keys    []string    `json:"keys"`
	Options CallOptions `json:"options"`
}

func (c *LocalConfig) TryOptions() *CallOptions {
	zeroOptions := Options{}
	if c.Options.Default == zeroOptions && len(c.Options.Overrides) == 0 {
		return nil
	}

	return &c.Options
}

type CallOptions struct {
	Default   Options           `json:"default"`
	Overrides []OptionOverrides `json:"overrides"`
}

type Options struct {
	// Allows calls without any form of authentication
	AllowAnonymous bool `json:"allow_anonymous" yaml:"allow_anonymous"`
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
