package config

import (
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

func UseJSONTags(dc *mapstructure.DecoderConfig) { dc.TagName = "json" }

func WithSquash(dc *mapstructure.DecoderConfig) {
	dc.Squash = true
}

type Viper struct {
	*viper.Viper
}

func NewViper() Viper {
	v := viper.NewWithOptions(
		viper.EnvKeyReplacer(newReplacer()),
	)

	v.SetConfigType("yaml")

	return Viper{v}
}

func (v Viper) Unmarshal(rawVal any) error {
	return v.Viper.Unmarshal(rawVal, func(dc *mapstructure.DecoderConfig) { dc.TagName = "json" })
}

func (v Viper) SetDefaults(c Section, prefix ...string) {
	var p string

	if len(prefix) > 0 {
		p = strings.Join(prefix, ".") + "."
	}

	for key, value := range c.Defaults() {
		v.SetDefault(p+key, value)
	}
}

func (v Viper) ReadConfig(r io.Reader) error {
	cfgBody, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	cfgText := substitueEnvVars(string(cfgBody))

	return v.Viper.ReadConfig(strings.NewReader(cfgText))
}

var (
	envRegex     = regexp.MustCompile(`(?U:\${.*})`)
	minVarExpLen = len("${}") + 1
)

func substitueEnvVars(s string) string {
	return envRegex.ReplaceAllStringFunc(strings.ReplaceAll(s, `"`, `'`), func(s string) string {
		// Trim off the '${' and '}'
		if len(s) < minVarExpLen {
			// This should never happen..
			return ""
		}

		varName := s[2 : len(s)-1]

		// Lookup the variable in the environment. We play by
		// bash rules.. if its undefined we'll treat it as an
		// empty string instead of raising an error.
		return os.Getenv(varName)
	})
}

type replacer struct {
	r *strings.Replacer
}

func newReplacer() *replacer {
	return &replacer{r: strings.NewReplacer(".", "_")}
}

func (r replacer) Replace(s string) string {
	if s == "TOPAZ_VERSION" {
		// Prevent the `version` field from be overridden by env vars.
		return ""
	}

	return r.r.Replace(s)
}
