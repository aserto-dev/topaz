package authentication

import (
	"strings"

	"github.com/rs/zerolog/log"
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
	Enabled  bool                   `json:"enabled"`
	Plugin   string                 `json:"plugin"`
	Settings map[string]interface{} `json:"settings,omitempty"`
}

// plugin: local - local authentication implementation.
type LocalSettings struct {
	APIKeys map[string]string `json:"api_keys"`
	Options CallOptions       `json:"options"`
	Keys    []string          `json:"keys"`
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

func (c *LocalSettings) transposeKeys() {
	if len(c.APIKeys) != 0 {
		log.Warn().Msg("config: auth.api_keys is deprecated, please use auth.keys")
	} else if c.APIKeys == nil {
		c.APIKeys = make(map[string]string)
	}

	for _, apiKey := range c.Keys {
		c.APIKeys[apiKey] = ""
	}
}
