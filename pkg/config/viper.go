package config

import (
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

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
